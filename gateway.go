package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/joho/godotenv"
)

func runService(name, dir, command string, args ...string) {
	cmd := exec.Command(command, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	log.Printf("🚀 Starting %s Service...", name)
	if err := cmd.Run(); err != nil {
		log.Printf("❌ %s Service exited with error: %v", name, err)
	}
}

// wsProxy hijacks the HTTP connection and pipes raw TCP to the backend
// This is required because httputil.ReverseProxy doesn't support WebSocket upgrades
func wsProxy(w http.ResponseWriter, r *http.Request, backendHost string) {
	// Connect to the backend WebSocket server
	backendConn, err := net.Dial("tcp", backendHost)
	if err != nil {
		http.Error(w, "Backend unreachable", http.StatusBadGateway)
		log.Printf("❌ WS proxy dial error to %s: %v", backendHost, err)
		return
	}
	defer backendConn.Close()

	// Hijack the client connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijack not supported", http.StatusInternalServerError)
		return
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, "Hijack failed", http.StatusInternalServerError)
		log.Printf("❌ Hijack error: %v", err)
		return
	}
	defer clientConn.Close()

	// Manually construct the HTTP upgrade request with proper Host for the backend
	reqURI := r.URL.RequestURI()
	raw := fmt.Sprintf("GET %s HTTP/1.1\r\n", reqURI)
	raw += fmt.Sprintf("Host: %s\r\n", backendHost)

	// Forward all original headers except Host
	for key, vals := range r.Header {
		if strings.EqualFold(key, "Host") {
			continue
		}
		for _, val := range vals {
			raw += fmt.Sprintf("%s: %s\r\n", key, val)
		}
	}
	raw += "\r\n"

	_, err = backendConn.Write([]byte(raw))
	if err != nil {
		log.Printf("❌ WS proxy write error: %v", err)
		return
	}

	log.Printf("✅ WebSocket tunnel established: %s -> %s", r.URL.Path, backendHost)

	// Pipe data bidirectionally
	errc := make(chan error, 2)

	go func() {
		_, err := io.Copy(backendConn, clientConn)
		errc <- err
	}()

	go func() {
		_, err := io.Copy(clientConn, backendConn)
		errc <- err
	}()

	// Wait for either direction to finish
	<-errc
}

func isWebSocketUpgrade(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Upgrade"), "websocket")
}

func proxyHandler() http.Handler {
	// ML Service Config
	mlHost := os.Getenv("ML_API_HOST")
	if mlHost == "" {
		mlHost = "127.0.0.1"
	}
	mlPort := os.Getenv("ML_API_PORT")
	if mlPort == "" {
		mlPort = "8001"
	}
	mlBackend := mlHost + ":" + mlPort

	authURL, _ := url.Parse("http://127.0.0.1:8080")
	cameraURL, _ := url.Parse("http://127.0.0.1:3001")

	authProxy := httputil.NewSingleHostReverseProxy(authURL)
	cameraProxy := httputil.NewSingleHostReverseProxy(cameraURL)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Proxying request: %s %s (Upgrade: %s)", r.Method, r.URL.Path, r.Header.Get("Upgrade"))

		// WebSocket routes — use raw TCP proxy
		if isWebSocketUpgrade(r) {
			if strings.HasPrefix(r.URL.Path, "/ws/stats") {
				wsProxy(w, r, mlBackend)
				return
			}
			if strings.HasPrefix(r.URL.Path, "/ws/chat") || strings.HasPrefix(r.URL.Path, "/ws/notifications") {
				wsProxy(w, r, "127.0.0.1:8083")
				return
			}
			// Fallback WS
			log.Printf("⚠️ Unknown WebSocket path: %s", r.URL.Path)
			http.Error(w, "Unknown WebSocket path", http.StatusNotFound)
			return
		}

		// Regular HTTP routes — use standard reverse proxy
		if strings.HasPrefix(r.URL.Path, "/api/v0/cctv") {
			cameraProxy.ServeHTTP(w, r)
			return
		}

		authProxy.ServeHTTP(w, r)
	})
}

func main() {
	// Load environment variables from .env
	_ = godotenv.Load()

	var wg sync.WaitGroup
	wg.Add(4) // 3 services + 1 proxy

	// Auth Service
	go func() {
		defer wg.Done()
		runService("Auth", "./Auth", "go", "run", "main.go")
	}()

	// Feedback Service
	go func() {
		defer wg.Done()
		runService("Feedback", "./Feedback", "go", "run", "main.go")
	}()

	// Camera Service
	go func() {
		defer wg.Done()
		runService("Camera", "./camera", "go", "run", "main.go")
	}()

	// Gateway Reverse Proxy
	go func() {
		defer wg.Done()
		log.Println("🌐 [INIT] Gateway Proxy Listening on :9999...")
		if err := http.ListenAndServe(":9999", proxyHandler()); err != nil {
			log.Fatalf("❌ Gateway proxy error: %v", err)
		}
	}()

	wg.Wait()
}