package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"tappyone/internal/utils"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin in development
		// In production, you should check the origin
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client represents a WebSocket client
type Client struct {
	ID     string
	UserID string
	Conn   *websocket.Conn
	Send   chan []byte
	Hub    *Hub
}

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	Clients map[*Client]bool

	// Inbound messages from the clients
	Broadcast chan []byte

	// Register requests from the clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client

	// User-specific channels
	UserChannels map[string]map[*Client]bool

	// Mutex for thread safety
	mutex sync.RWMutex
}

// Message types for WebSocket communication
type WSMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	UserID    string      `json:"user_id,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// Message types
const (
	MessageTypeNewMessage    = "new_message"
	MessageTypeMessageStatus = "message_status"
	MessageTypeTyping        = "typing"
	MessageTypePresence      = "presence"
	MessageTypeConnection    = "connection"
	MessageTypeError         = "error"
	MessageTypePing          = "ping"
	MessageTypePong          = "pong"
)

// Global hub instance
var wsHub *Hub

// Initialize WebSocket hub
func InitWebSocketHub() {
	wsHub = &Hub{
		Clients:      make(map[*Client]bool),
		Broadcast:    make(chan []byte),
		Register:     make(chan *Client),
		Unregister:   make(chan *Client),
		UserChannels: make(map[string]map[*Client]bool),
	}
	go wsHub.Run()
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mutex.Lock()
			h.Clients[client] = true
			
			// Add to user-specific channel
			if h.UserChannels[client.UserID] == nil {
				h.UserChannels[client.UserID] = make(map[*Client]bool)
			}
			h.UserChannels[client.UserID][client] = true
			
			h.mutex.Unlock()
			
			log.Printf("Client %s connected for user %s", client.ID, client.UserID)
			
			// Send connection confirmation
			message := WSMessage{
				Type:      MessageTypeConnection,
				Data:      map[string]string{"status": "connected"},
				Timestamp: time.Now(),
			}
			if data, err := json.Marshal(message); err == nil {
				select {
				case client.Send <- data:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}

		case client := <-h.Unregister:
			h.mutex.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				
				// Remove from user-specific channel
				if userClients, exists := h.UserChannels[client.UserID]; exists {
					delete(userClients, client)
					if len(userClients) == 0 {
						delete(h.UserChannels, client.UserID)
					}
				}
				
				close(client.Send)
				log.Printf("Client %s disconnected for user %s", client.ID, client.UserID)
			}
			h.mutex.Unlock()

		case message := <-h.Broadcast:
			h.mutex.RLock()
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// BroadcastToUser sends a message to all connections of a specific user
func (h *Hub) BroadcastToUser(userID string, message WSMessage) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	h.mutex.RLock()
	userClients, exists := h.UserChannels[userID]
	h.mutex.RUnlock()

	if !exists {
		return
	}

	for client := range userClients {
		select {
		case client.Send <- data:
		default:
			close(client.Send)
			h.mutex.Lock()
			delete(h.Clients, client)
			delete(userClients, client)
			h.mutex.Unlock()
		}
	}
}

// NewWebSocketHandler creates a WebSocket handler with JWT auth via query param
func NewWebSocketHandler(authService interface{}, jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("[WEBSOCKET] New connection attempt")
		
		// Get token from query parameter
		token := c.Query("token")
		log.Printf("[WEBSOCKET] Token received: %s", token[:50]+"...")
		
		if token == "" {
			log.Printf("[WEBSOCKET] No token provided")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
			return
		}

		log.Printf("[WEBSOCKET] JWT Secret length: %d", len(jwtSecret))
		log.Printf("[WEBSOCKET] JWT Secret preview: %s", jwtSecret[:10]+"...")
		
		// Validate token and get user ID
		userID, err := utils.ValidateJWTTokenWithSecret(token, jwtSecret)
		if err != nil {
			log.Printf("[WEBSOCKET] Token validation failed: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token", "details": err.Error()})
			return
		}

		log.Printf("[WEBSOCKET] Token validated successfully for user: %s", userID)
		
		// Store user ID in context for HandleWebSocket
		c.Set("userID", userID)
		HandleWebSocket(c)
	}
}

// WebSocket handler
func HandleWebSocket(c *gin.Context) {
	// Get user from context (set by NewWebSocketHandler)
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	log.Printf("[WEBSOCKET] Attempting to upgrade connection for user: %s", userID)
	
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("[WEBSOCKET] Upgrade error: %v", err)
		log.Printf("[WEBSOCKET] Request headers: %+v", c.Request.Header)
		return
	}
	
	log.Printf("[WEBSOCKET] Connection upgraded successfully for user: %s", userID)

	// Create client
	client := &Client{
		ID:     generateClientID(),
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Hub:    wsHub,
	}

	log.Printf("[WEBSOCKET] Registering client %s for user %s", client.ID, client.UserID)
	
	// Register client
	client.Hub.Register <- client

	// Start goroutines for reading and writing
	log.Printf("[WEBSOCKET] Starting read/write pumps for client %s", client.ID)
	go client.writePump()
	go client.readPump()
}

// Read messages from WebSocket
func (c *Client) readPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	// Set read deadline and pong handler
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming messages
		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		// Handle different message types
		switch wsMsg.Type {
		case MessageTypePing:
			// Respond with pong
			pongMsg := WSMessage{
				Type:      MessageTypePong,
				Data:      map[string]string{"status": "pong"},
				Timestamp: time.Now(),
			}
			if data, err := json.Marshal(pongMsg); err == nil {
				select {
				case c.Send <- data:
				default:
					return
				}
			}
		case MessageTypeTyping:
			// Broadcast typing status to other users
			// This would be implemented based on your business logic
			log.Printf("User %s is typing", c.UserID)
		}
	}
}

// Write messages to WebSocket
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Generate unique client ID
func generateClientID() string {
	return time.Now().Format("20060102150405") + "-" + utils.GenerateRandomString(8)
}

// GetWebSocketHub returns the global WebSocket hub
func GetWebSocketHub() *Hub {
	return wsHub
}

// BroadcastNewMessage sends a new WhatsApp message to all user connections
func BroadcastNewMessage(userID string, message interface{}) {
	if wsHub == nil {
		log.Printf("[WEBSOCKET] Hub not initialized")
		return
	}

	wsMessage := WSMessage{
		Type:      MessageTypeNewMessage,
		Data:      message,
		UserID:    userID,
		Timestamp: time.Now(),
	}

	log.Printf("[WEBSOCKET] Broadcasting new message to user %s", userID)
	wsHub.BroadcastToUser(userID, wsMessage)
}
