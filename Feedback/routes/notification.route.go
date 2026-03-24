package routes

import (
	"Feedback/db"
	"Feedback/utils"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// --- NOTIFICATION TYPES ---
type NotifMsg struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Type    string `json:"type"` // info, error
}

type NotifClient struct {
	Conn   *websocket.Conn
	Send   chan []byte
	UserID string
}

type NotifHub struct {
	Clients    map[*NotifClient]bool
	UserIndex  map[string]*NotifClient // MAP: UserID -> Client
	Register   chan *NotifClient
	Unregister chan *NotifClient
	Mu         sync.Mutex
}

var N_Hub = &NotifHub{
	Clients:    make(map[*NotifClient]bool),
	UserIndex:  make(map[string]*NotifClient),
	Register:   make(chan *NotifClient),
	Unregister: make(chan *NotifClient),
}

func (h *NotifHub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Mu.Lock()
			h.Clients[client] = true
			h.UserIndex[client.UserID] = client // Index User
			h.Mu.Unlock()
		case client := <-h.Unregister:
			h.Mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				delete(h.UserIndex, client.UserID) // Remove User
				close(client.Send)
			}
			h.Mu.Unlock()
		}
	}
}

// --- INTERNAL TRIGGER (Called by other microservices) ---
func TriggerNotificationHandler(w http.ResponseWriter, r *http.Request) {
	type Req struct {
		TargetID string `json:"target_id"`
		Content  string `json:"content"`
	}
	var req Req
	json.NewDecoder(r.Body).Decode(&req)

	// Send to specific user
	N_Hub.Mu.Lock()
	if client, ok := N_Hub.UserIndex[req.TargetID]; ok {
		msg, _ := json.Marshal(NotifMsg{Title: "Alert", Content: req.Content, Type: "info"})
		client.Send <- msg
	}
	N_Hub.Mu.Unlock()

	w.Write([]byte(`{"status":"sent"}`))
}

// --- WS HANDLER ---
var NotifUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NotificationHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	claims, err := utils.ParseToken(token)
	if err != nil { http.Error(w, "Unauthorized", 401); return }

	// 🔒 Validate Login Status
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

	conn, _ := NotifUpgrader.Upgrade(w, r, nil)
	client := &NotifClient{Conn: conn, Send: make(chan []byte, 256), UserID: claims["user_id"].(string)}
	
	N_Hub.Register <- client

	// Write Pump
	go func() {
		defer conn.Close()
		for msg := range client.Send { conn.WriteMessage(websocket.TextMessage, msg) }
	}()

	// Read Pump (Notifications are usually Read-Only for client, but keep alive needs read)
	go func() {
		defer func() { N_Hub.Unregister <- client; conn.Close() }()
		for {
			if _, _, err := conn.ReadMessage(); err != nil { break }
		}
	}()
}