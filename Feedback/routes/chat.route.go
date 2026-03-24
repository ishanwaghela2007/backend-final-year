package routes

import (
	"Feedback/utils"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"Feedback/db"
	"time"

	"github.com/gocql/gocql"
	"github.com/gorilla/websocket"
)

// --- CHAT TYPES ---
type ChatMsg struct {
	Type    string `json:"type"`    // "msg" or "cmd" (open/close)
	Content string `json:"content"`
	Sender  string `json:"sender"`  // UserID
	Role    string `json:"role"`    // Admin/Staff
}

type ChatClient struct {
	Conn *websocket.Conn
	Send chan []byte
	Role string
}

type ChatHub struct {
	Clients     map[*ChatClient]bool
	Broadcast   chan ChatMsg
	Register    chan *ChatClient
	Unregister  chan *ChatClient
	ChatAllowed bool
	Mu          sync.Mutex
}

var C_Hub = &ChatHub{
	Clients:     make(map[*ChatClient]bool),
	Broadcast:   make(chan ChatMsg),
	Register:    make(chan *ChatClient),
	Unregister:  make(chan *ChatClient),
	ChatAllowed: false, // Default closed
}

func (h *ChatHub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Mu.Lock()
			h.Clients[client] = true
			
			// Send current chat permission status to the new client
			statusMsg := "close"
			if h.ChatAllowed {
				statusMsg = "open"
			}
			msg := ChatMsg{
				Type:    "cmd",
				Content: statusMsg,
				Role:    "system",
			}
			bytes, _ := json.Marshal(msg)
			client.Send <- bytes
			
			h.Mu.Unlock()
		case client := <-h.Unregister:
			h.Mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
			h.Mu.Unlock()
		case msg := <-h.Broadcast:
			h.Mu.Lock()
			
			// 1. Logic: Admin Commands
			if msg.Role == "admin" && msg.Type == "cmd" {
				if msg.Content == "open" { h.ChatAllowed = true }
				if msg.Content == "close" { h.ChatAllowed = false }
			}

			// 2. Logic: Permission Check
			shouldSend := true
			if msg.Role == "staff" && !h.ChatAllowed {
				shouldSend = false
			}

			// 3. Logic: Broadcast
			if shouldSend {
				bytes, _ := json.Marshal(msg)
				for client := range h.Clients {
					// Send to everyone (or add specific filtering here)
					select {
					case client.Send <- bytes:
					default:
						close(client.Send); delete(h.Clients, client)
					}
				}
			}
			h.Mu.Unlock()
		}
	}
}

// --- HANDLER ---
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ChatHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	claims, err := utils.ParseToken(token)
	if err != nil { http.Error(w, "Unauthorized", 401); return }

	// 🔒 Validate Login Status (Cross-Keyspace Query)
	userID := claims["user_id"].(string)
	var isLoggedIn bool
	if err := db.Session.Query(`SELECT isloggedin FROM auth.users WHERE id = ?`, userID).Scan(&isLoggedIn); err != nil {
		http.Error(w, "User not found", 401)
		return
	}
	if !isLoggedIn {
		http.Error(w, "User is not logged in", 403)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ WebSocket upgrade failed: %v", err)
		return
	}
	client := &ChatClient{Conn: conn, Send: make(chan []byte, 256), Role: claims["role"].(string)}
	
	C_Hub.Register <- client

	// Write Pump (Inline for brevity)
	go func() {
		defer conn.Close()
		for msg := range client.Send { conn.WriteMessage(websocket.TextMessage, msg) }
	}()

	// Read Pump
	go func() {
		defer func() { C_Hub.Unregister <- client; conn.Close() }()
		for {
			_, bytes, err := conn.ReadMessage()
			if err != nil { break }
			var msg ChatMsg
			if json.Unmarshal(bytes, &msg) == nil {
				msg.Sender = claims["user_id"].(string)
				msg.Role = claims["role"].(string)

				// Async Save to DB
				go func(m ChatMsg) {
					// Use specific channel ID logic if needed, defaulting to "global"
					channelID := "global"
					logID := gocql.TimeUUID()
					
					query := `INSERT INTO messages (channel_id, id, sender_id, sender_role, content, created_at) VALUES (?, ?, ?, ?, ?, ?)`
					if err := db.Session.Query(query, channelID, logID, m.Sender, m.Role, m.Content, time.Now()).Exec(); err != nil {
						// log error but don't stop broadcast
						// fmt.Println("Error saving message:", err)
					}
				}(msg)

				C_Hub.Broadcast <- msg
			}
		}
	}()
}