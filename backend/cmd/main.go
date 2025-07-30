package main

import (
	"log"
	"t-9/internal/api"
)

func main() {
	r := api.SetupRoutes()
	
	log.Println("Starting T-9 server on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}