package ws

import (
	"crypto/rand"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"t-9/internal/config"
	"t-9/internal/game"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

const (
	maxRooms         = 100               // Maximum concurrent game rooms
	roomTimeout      = 30 * time.Minute  // Delete inactive rooms after 30 minutes
	cleanupInterval  = 5 * time.Minute   // Run cleanup every 5 minutes
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Get the Origin header
		origin := r.Header.Get("Origin")
		if origin == "" {
			// No Origin header, likely same-origin request
			return true
		}

		// In production, validate against specific origins
		if isProduction() {
			return isValidOrigin(origin)
		}

		// Development: allow localhost and common dev origins
		return isDevelopmentOrigin(origin)
	},
}

// isProduction checks if we're running in production
func isProduction() bool {
	// Check environment variable, could be expanded with other checks
	return getEnv("ENVIRONMENT", "development") == "production"
}

// isValidOrigin validates origin against allowed production origins
func isValidOrigin(origin string) bool {
	allowedOrigins := config.DefaultConfig.CORS.AllowedOrigins
	
	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
	}
	return false
}

// isDevelopmentOrigin allows common development origins
func isDevelopmentOrigin(origin string) bool {
	allowedDevOrigins := []string{
		"http://localhost:5173",
		"http://localhost:4173",
		"http://127.0.0.1:5173",
		"http://127.0.0.1:4173",
	}
	
	for _, allowed := range allowedDevOrigins {
		if origin == allowed {
			return true
		}
	}
	return false
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Hub manages all game rooms and WebSocket connections
type Hub struct {
	rooms      map[string]*GameRoom
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
	mu         sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]*GameRoom),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan []byte),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	// Start cleanup goroutine
	ticker := time.NewTicker(cleanupInterval)
	go func() {
		for range ticker.C {
			h.cleanupInactiveRooms()
		}
	}()

	for {
		select {
		case client := <-h.register:
			h.handleRegister(client)

		case client := <-h.unregister:
			h.handleUnregister(client)
		}
	}
}

// ServeWS handles WebSocket connections
func (h *Hub) ServeWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	client := &Client{
		ID:   generateClientID(),
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	h.register <- client

	// Start goroutines for reading and writing
	go h.writePump(client)
	go h.readPump(client)
}

// handleRegister registers a new client
func (h *Hub) handleRegister(client *Client) {
	log.Printf("Client %s connected", client.ID)
}

// handleUnregister removes a client
func (h *Hub) handleUnregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if client.GameID != "" {
		if room, exists := h.rooms[client.GameID]; exists {
			h.removeClientFromRoom(room, client)
		}
	}

	close(client.Send)
	client.Conn.Close()
	log.Printf("Client %s disconnected", client.ID)
}

// removeClientFromRoom removes a client from a game room
func (h *Hub) removeClientFromRoom(room *GameRoom, client *Client) {
	delete(room.Clients, client.ID)

	if room.PlayerX == client {
		room.PlayerX = nil
	}
	if room.PlayerO == client {
		room.PlayerO = nil
	}

	// Notify other clients
	msg := Message{
		Type:    MsgTypePlayerLeave,
		Player:  client.Player,
		Message: "Player disconnected",
	}
	h.broadcastToRoom(room, msg)

	// Remove room if empty
	if len(room.Clients) == 0 {
		delete(h.rooms, room.ID)
		log.Printf("Room %s deleted (empty)", room.ID)
	}
}

// readPump handles incoming messages from clients
func (h *Hub) readPump(client *Client) {
	defer func() {
		h.unregister <- client
	}()

	for {
		var msg Message
		err := client.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		h.handleMessage(client, msg)
	}
}

// writePump handles outgoing messages to clients
func (h *Hub) writePump(client *Client) {
	defer client.Conn.Close()

	for {
		select {
		case message, ok := <-client.Send:
			if !ok {
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			client.Conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

// handleMessage processes incoming messages
func (h *Hub) handleMessage(client *Client, msg Message) {
	switch msg.Type {
	case MsgTypeJoinGame:
		h.handleJoinGame(client, msg.GameID)
	case MsgTypeMove:
		if msg.Move != nil {
			h.handleMove(client, *msg.Move)
		}
	}
}

// handleJoinGame handles a client joining a game
func (h *Hub) handleJoinGame(client *Client, gameID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	room, exists := h.rooms[gameID]
	if !exists {
		// Check if we've hit max rooms limit
		if len(h.rooms) >= maxRooms {
			h.sendError(client, "Server is at capacity. Please try again later.")
			log.Printf("Max rooms reached (%d), rejecting new game", maxRooms)
			return
		}

		// Create new room
		room = &GameRoom{
			ID:           gameID,
			Game:         game.NewGame(),
			Clients:      make(map[string]*Client),
			LastActivity: time.Now(),
		}
		h.rooms[gameID] = room
	}

	// Update activity timestamp
	room.LastActivity = time.Now()

	// Check if room is full
	if len(room.Clients) >= 2 {
		h.sendError(client, "Game room is full")
		return
	}

	// Assign player
	if room.PlayerX == nil {
		client.Player = game.X
		room.PlayerX = client
	} else if room.PlayerO == nil {
		client.Player = game.O
		room.PlayerO = client
	}

	client.GameID = gameID
	room.Clients[client.ID] = client

	// Send game state to joining player
	h.sendGameState(client, room.Game)

	// Notify other players
	msg := Message{
		Type:    MsgTypePlayerJoin,
		Player:  client.Player,
		Message: "Player joined",
	}
	h.broadcastToRoom(room, msg)

	log.Printf("Client %s joined game %s as player %v", client.ID, gameID, client.Player)
}

// handleMove handles a game move
func (h *Hub) handleMove(client *Client, move game.Move) {
	h.mu.RLock()
	room, exists := h.rooms[client.GameID]
	h.mu.RUnlock()

	if !exists {
		h.sendError(client, "Game not found")
		return
	}

	// Validate it's the player's turn
	if room.Game.CurrentPlayer != client.Player {
		h.sendError(client, "Not your turn")
		return
	}

	// Make the move
	if err := room.Game.MakeMove(move); err != nil {
		h.sendError(client, err.Error())
		return
	}

	// Update room activity
	h.mu.Lock()
	room.LastActivity = time.Now()
	h.mu.Unlock()

	// Broadcast updated game state to all players in room
	msg := Message{
		Type: MsgTypeGameState,
		Game: room.Game,
	}
	h.broadcastToRoom(room, msg)
}

// sendGameState sends current game state to a client
func (h *Hub) sendGameState(client *Client, gameState *game.GameState) {
	msg := Message{
		Type: MsgTypeGameState,
		Game: gameState,
	}
	h.sendToClient(client, msg)
}

// sendError sends an error message to a client
func (h *Hub) sendError(client *Client, errorMsg string) {
	msg := Message{
		Type:  MsgTypeError,
		Error: errorMsg,
	}
	h.sendToClient(client, msg)
}

// sendToClient sends a message to a specific client
func (h *Hub) sendToClient(client *Client, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	select {
	case client.Send <- data:
	default:
		close(client.Send)
		delete(h.rooms[client.GameID].Clients, client.ID)
	}
}

// broadcastToRoom sends a message to all clients in a room
func (h *Hub) broadcastToRoom(room *GameRoom, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	for _, client := range room.Clients {
		select {
		case client.Send <- data:
		default:
			close(client.Send)
			delete(room.Clients, client.ID)
		}
	}
}

// generateClientID creates a unique client ID using secure random generation
func generateClientID() string {
	return "client_" + generateSecureString(8)
}

// generateSecureString generates a cryptographically secure random string
func generateSecureString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp if crypto/rand fails
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}

	// Convert random bytes to charset
	for i := range b {
		b[i] = charset[b[i]%byte(len(charset))]
	}

	return string(b) + "_" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

// cleanupInactiveRooms removes rooms that haven't had activity in a while
func (h *Hub) cleanupInactiveRooms() {
	h.mu.Lock()
	defer h.mu.Unlock()

	now := time.Now()
	deletedCount := 0

	for gameID, room := range h.rooms {
		if now.Sub(room.LastActivity) > roomTimeout {
			// Notify any remaining clients
			for _, client := range room.Clients {
				h.sendError(client, "Game closed due to inactivity")
				client.Conn.Close()
			}

			delete(h.rooms, gameID)
			deletedCount++
			log.Printf("Cleaned up inactive room: %s (inactive for %v)", gameID, now.Sub(room.LastActivity))
		}
	}

	if deletedCount > 0 {
		log.Printf("Cleanup complete: removed %d inactive rooms, %d rooms remaining", deletedCount, len(h.rooms))
	}
}