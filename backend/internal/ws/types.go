package ws

import (
	"t-9/internal/game"
	"time"

	"github.com/gorilla/websocket"
)

// Client represents a connected WebSocket client
type Client struct {
	ID       string
	Conn     *websocket.Conn
	GameID   string
	Player   game.Player
	Send     chan []byte
}

// GameRoom manages a multiplayer game session
type GameRoom struct {
	ID           string
	Game         *game.GameState
	Clients      map[string]*Client
	PlayerX      *Client
	PlayerO      *Client
	LastActivity time.Time
}

// Message types for WebSocket communication
type MessageType string

const (
	MsgTypeJoinGame    MessageType = "join_game"
	MsgTypeGameState   MessageType = "game_state"
	MsgTypeMove        MessageType = "move"
	MsgTypePlayerJoin  MessageType = "player_join"
	MsgTypePlayerLeave MessageType = "player_leave"
	MsgTypeError       MessageType = "error"
)

// WebSocket message structure
type Message struct {
	Type    MessageType `json:"type"`
	GameID  string      `json:"gameId,omitempty"`
	Move    *game.Move  `json:"move,omitempty"`
	Game    interface{} `json:"game,omitempty"`
	Player  game.Player `json:"player,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// JoinGameRequest represents a request to join a game
type JoinGameRequest struct {
	GameID string `json:"gameId"`
}