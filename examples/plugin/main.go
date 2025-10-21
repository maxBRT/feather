package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/maxBRT/feather"
)

// LoggingPluginFactory creates a simple request logging middleware.
// The factory read from plugin-specific configuration.
func LoggingPluginFactory(config map[string]any) feather.Middleware {
	// Example: optional configuration lookup
	prefix, _ := config["prefix"].(string)
	if prefix == "" {
		prefix = "[Feather]"
	}

	// Return the actual middleware function
	return func(next http.HandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			log.Printf("%s -> %s %s", prefix, r.Method, r.URL.Path)

			next.ServeHTTP(w, r)

			log.Printf("%s <- %s (%v)", prefix, r.URL.Path, time.Since(start))
		})
	}
}

func main() {
	gw, err := feather.New("config.yaml")
	if err != nil {
		log.Fatalf("failed to initialize Feather gateway: %v", err)
	}

	// Register your plugin with the gateway
	gw.RegisterPlugin("logger", LoggingPluginFactory)

	if err := gw.Run("8080"); err != nil {
		log.Printf("gateway exited with error: %v\n", err)
		os.Exit(1)
	}
}
