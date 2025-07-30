package main

import (
	"log"
	"t-9/internal/api"
	"t-9/internal/ws"
)

func main() {
	// Create WebSocket hub
	hub := ws.NewHub()
	go hub.Run()

	// Setup routes with WebSocket support
	r := api.SetupRoutes(hub)
	
	log.Println("Starting T-9 server with multiplayer support on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}