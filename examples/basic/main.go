package main

import (
	"log"
	"os"

	"github.com/maxBRT/feather"
)

func main() {
	// Initialize a new Feather Gateway instance from the provided YAML configuration file.
	// The configuration typically defines routes, backends, and possibly middleware settings.
	gw, err := feather.New("config.yaml")
	if err != nil {
		log.Fatalf("failed to initialize Feather: %v", err)
	}

	// Start the gateway on port 8080.
	// Feather will read the routes from config.yaml and begin forwarding requests accordingly.
	if err := gw.Run("8080"); err != nil {
		log.Printf("Feather exited with error: %v\n", err)
		os.Exit(1)
	}
}
