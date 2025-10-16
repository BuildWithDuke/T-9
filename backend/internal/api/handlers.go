package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"sync"
	"time"

	"t-9/internal/ai"
	"t-9/internal/config"
	"t-9/internal/game"
	"t-9/internal/logging"
	"t-9/internal/ws"

	"github.com/gin-gonic/gin"
)

// GameManager handles game sessions
type GameManager struct {
	games map[string]*game.GameState
	mu    sync.RWMutex
}

func NewGameManager() *GameManager {
	return &GameManager{
		games: make(map[string]*game.GameState),
	}
}

// CreateGame creates a new game session
func (gm *GameManager) CreateGame(c *gin.Context) {
	gameID := generateGameID()
	newGame := game.NewGame()
	
	gm.mu.Lock()
	gm.games[gameID] = newGame
	gm.mu.Unlock()

	logging.DefaultLogger.Info("Game created", map[string]interface{}{
		"gameId":     gameID,
		"clientIP":   c.ClientIP(),
		"userAgent":  c.Request.UserAgent(),
	})

	c.JSON(http.StatusCreated, gin.H{
		"gameId": gameID,
		"game":   newGame,
	})
}

// GetGame retrieves a game by ID
func (gm *GameManager) GetGame(c *gin.Context) {
	gameID := c.Param("id")
	
	gm.mu.RLock()
	gameState, exists := gm.games[gameID]
	gm.mu.RUnlock()
	
	if !exists {
		logging.DefaultLogger.Warning("Game not found", map[string]interface{}{
			"gameId":    gameID,
			"clientIP":  c.ClientIP(),
		})
		c.JSON(http.StatusNotFound, NewNotFoundError("Game"))
		return
	}

	logging.DefaultLogger.Info("Game retrieved", map[string]interface{}{
		"gameId":    gameID,
		"clientIP":  c.ClientIP(),
	})

	c.JSON(http.StatusOK, gameState)
}

// MakeMove handles a player's move
func (gm *GameManager) MakeMove(c *gin.Context) {
	gameID := c.Param("id")
	
	// Use read lock for initial lookup
	gm.mu.RLock()
	gameState, exists := gm.games[gameID]
	gm.mu.RUnlock()
	
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}

	var move game.Move
	if err := c.ShouldBindJSON(&move); err != nil {
		c.JSON(http.StatusBadRequest, NewInvalidInputError("Invalid move format: "+err.Error()))
		return
	}

	// Upgrade to write lock only for the actual move operation
	gm.mu.Lock()
	defer gm.mu.Unlock()
	
	// Re-check existence in case it was deleted
	gameState, exists = gm.games[gameID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}
	
	if err := gameState.MakeMove(move); err != nil {
		c.JSON(http.StatusBadRequest, NewGameLogicError(err.Error()))
		return
	}

	c.JSON(http.StatusOK, gameState)
}

// MakeAIMove handles AI moves for single-player games
func (gm *GameManager) MakeAIMove(c *gin.Context) {
	gameID := c.Param("id")
	
	gm.mu.RLock()
	gameState, exists := gm.games[gameID]
	gm.mu.RUnlock()
	
	if !exists {
		c.JSON(http.StatusNotFound, NewNotFoundError("Game"))
		return
	}

	var request struct {
		Difficulty string `json:"difficulty"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, NewInvalidInputError("Invalid request format: "+err.Error()))
		return
	}

	// Parse difficulty
	var difficulty ai.Difficulty
	switch request.Difficulty {
	case "easy":
		difficulty = ai.Easy
	case "medium":
		difficulty = ai.Medium
	case "hard":
		difficulty = ai.Hard
	default:
		difficulty = ai.Medium
	}

	// Check if game is over
	if gameState.GameOver {
		c.JSON(http.StatusBadRequest, NewGameLogicError("Game is already over"))
		return
	}

	// Create AI player for current player
	aiPlayer := ai.NewAIPlayer(difficulty, gameState.CurrentPlayer)
	
	// Get AI move
	aiMove := aiPlayer.GetBestMove(gameState)
	
	// Check if AI found a valid move (empty move detection)
	if aiMove.Player == 0 {
		c.JSON(http.StatusBadRequest, NewGameLogicError("No valid moves available for AI"))
		return
	}
	
	// Validate and make the move
	if err := gameState.MakeMove(aiMove); err != nil {
		c.JSON(http.StatusBadRequest, NewGameLogicError("AI move validation failed: "+err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"game": gameState,
		"move": aiMove,
	})
}

// generateGameID creates a unique game ID using secure random generation
func generateGameID() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp if crypto/rand fails
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return hex.EncodeToString(bytes)
}

var gameManager *GameManager

// HealthCheck returns server health status
func HealthCheck(c *gin.Context) {
	logging.DefaultLogger.Info("Health check requested", map[string]interface{}{
		"clientIP": c.ClientIP(),
	})

	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"timestamp": time.Now(),
		"version": "1.0.0",
	})
}

func SetupRoutes(hub *ws.Hub) *gin.Engine {
	gameManager = NewGameManager()
	r := gin.Default()
	
	// Set trusted proxies to avoid warning
	r.SetTrustedProxies([]string{"127.0.0.1", "::1"})

	// Enable CORS for frontend
	r.Use(func(c *gin.Context) {
		// Use more restrictive CORS settings from config
		allowedOrigins := config.DefaultConfig.CORS.AllowedOrigins
		origin := c.Request.Header.Get("Origin")
		
		// Allow requests with no Origin (like curl)
		if origin == "" {
			c.Next()
			return
		}
		
		// Check if origin is allowed
		for _, allowed := range allowedOrigins {
			if origin == allowed {
				c.Header("Access-Control-Allow-Origin", origin)
				break
			}
		}
		
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// Health check endpoint
	r.GET("/health", HealthCheck)

	// WebSocket endpoint
	r.GET("/ws", hub.ServeWS)

	api := r.Group("/api/v1")
	{
		api.POST("/games", gameManager.CreateGame)
		api.GET("/games/:id", gameManager.GetGame)
		api.POST("/games/:id/moves", gameManager.MakeMove)
		api.POST("/games/:id/ai-move", gameManager.MakeAIMove)
		api.GET("/health", HealthCheck)
	}

	return r
}