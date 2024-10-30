package main

import (
	"log"
	"net/http"
	"rule-engine/bl"
)

const port = "8081"

func main() {
	// Load rules on startup
	if err := bl.LoadRules(); err != nil {
		log.Fatalf("Failed to load rules: %v", err)
	}

	// Define the HTTP handler function
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		bl.HandleRequest(w, r)
	})
	// Start the HTTP server
	log.Printf("Server running on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
