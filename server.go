package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// StartServer starts the server mode with the given configuration
func StartServer(config *Config, processTask func() ([]string, error)) {
	log.Printf("Starting server on port %d\n", config.Port)
	log.Printf("Watching directory: %s\n", config.Directory)
	log.Printf("Using delimiter: '%s'\n", config.Delimiter)
	log.Printf("Using format string: '%s'\n", config.FormatString)

	// Setup webhook endpoint
	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		log.Println("Received webhook notification, processing files...")

		// Process files in the directory
		processedFiles, err := processTask()

		if err != nil {
			log.Printf("Error processing files: %v\n", err)
			http.Error(w, fmt.Sprintf("Error processing files: %v", err), http.StatusInternalServerError)
			return
		}

		// Return the list of processed files
		response := map[string]interface{}{
			"status":         "success",
			"processedFiles": processedFiles,
			"timestamp":      time.Now().Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Setup status endpoint
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		status := map[string]interface{}{
			"status":     "running",
			"directory":  config.Directory,
			"startedAt":  time.Now().Format(time.RFC3339),
			"delimiter":  config.Delimiter,
			"formatSpec": config.FormatString,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	})

	// Start HTTP server
	serverAddr := fmt.Sprintf(":%d", config.Port)

	if config.CertFile != "" && config.KeyFile != "" {
		err := http.ListenAndServeTLS(serverAddr, ensureAbsolutePath(config.CertFile), ensureAbsolutePath(config.KeyFile), nil)
		if err == nil {
			return
		}
		fmt.Printf("SSL Certificate file error! %v", err)
	}
	log.Fatal(http.ListenAndServe(serverAddr, nil))

}
