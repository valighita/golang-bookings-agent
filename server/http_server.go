package server

import (
	"log"
	"net/http"
	"os"
	"valighita/bookings-ai-agent/agent"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func handleWebSocket(agentFactory agent.AgentFactory) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("Error upgrading to websocket:", err)
			return
		}
		defer conn.Close()

		// Create agent
		agent, err := agentFactory.CreateAgent()
		if err != nil {
			log.Println("Error creating agent:", err)
			return
		}

		for {
			// Read message from browser
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("Error reading message:", err)
				break
			}

			// Process the message (this is where you would integrate with your agent)
			response, err := agent.GetCompletion(string(msg))
			if err != nil {
				log.Println("Error getting completion:", err)
				break
			}

			// Write message back to browser
			err = conn.WriteMessage(websocket.TextMessage, []byte(response))
			if err != nil {
				log.Println("Error writing message:", err)
				break
			}
		}
	}
}

func basicAuth(next http.Handler, username, password string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != username || pass != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RunHttpServer(agentFactory agent.AgentFactory) {
	port := os.Getenv("HTTP_SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	username := os.Getenv("HTTP_SERVER_USERNAME")
	password := os.Getenv("HTTP_SERVER_PASSWORD")
	withAuth := username != "" && password != ""

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	if withAuth {
		r.Use(func(next http.Handler) http.Handler {
			return basicAuth(next, username, password)
		})
	}

	// Define WebSocket route
	r.Get("/ws", handleWebSocket(agentFactory))

	// serve frontend/index.html on /
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/index.html")
	})
	r.Get("/style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "frontend/style.css")
	})

	log.Printf("Starting server on port %s\n", port)
	err := http.ListenAndServe(":"+port, r)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
