package api

import (
	"net/http"
	"strconv"

	"t-9/internal/game"

	"github.com/gin-gonic/gin"
)

// GameManager handles game sessions
type GameManager struct {
	games map[string]*game.GameState
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
	gm.games[gameID] = newGame

	c.JSON(http.StatusCreated, gin.H{
		"gameId": gameID,
		"game":   newGame,
	})
}

// GetGame retrieves a game by ID
func (gm *GameManager) GetGame(c *gin.Context) {
	gameID := c.Param("id")
	gameState, exists := gm.games[gameID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}

	c.JSON(http.StatusOK, gameState)
}

// MakeMove handles a player's move
func (gm *GameManager) MakeMove(c *gin.Context) {
	gameID := c.Param("id")
	gameState, exists := gm.games[gameID]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}

	var move game.Move
	if err := c.ShouldBindJSON(&move); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid move format"})
		return
	}

	if err := gameState.MakeMove(move); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gameState)
}

// generateGameID creates a unique game ID
func generateGameID() string {
	// Simple implementation - in production, use UUID or better method
	return strconv.FormatInt(int64(len(gameManager.games)+1), 36)
}

var gameManager *GameManager

func SetupRoutes() *gin.Engine {
	gameManager = NewGameManager()
	r := gin.Default()

	// Enable CORS for frontend
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	api := r.Group("/api/v1")
	{
		api.POST("/games", gameManager.CreateGame)
		api.GET("/games/:id", gameManager.GetGame)
		api.POST("/games/:id/moves", gameManager.MakeMove)
	}

	return r
}