package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// the structure of the data we expect in motivePipeline POST request
type Token struct {
	Value string `json:"token"`
}

var events = map[string][]string{
	"vehicles": {"https://eoww187fd6vl0sa.m.pipedream.net"},
	"drivers":  {"https://eoww187fd6vl0sa.m.pipedream.net"},
	"trailers": {"https://eoww187fd6vl0sa.m.pipedream.net"},
}

func main() {
	// Define the routes
	http.HandleFunc("/", home)
	http.HandleFunc("/motive-pipeline", motivePipeline)

	// Start the server on port 8080
	fmt.Println("Server is listening on port 8080...")
	http.ListenAndServe(":8080", nil)
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{"status": "ok", "msg": "we're up"}
	json.NewEncoder(w).Encode(response)
}

func motivePipeline(w http.ResponseWriter, r *http.Request) {
	// Check if the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// check if the json payload is correctly formed
	var newToken Token
	err := json.NewDecoder(r.Body).Decode(&newToken)
	if err != nil {
		http.Error(w, "Error decoding JSON payload", http.StatusBadRequest)
		return
	}

	// check if the json payload has the expected field (token)
	if newToken.Value == "" {
		http.Error(w, "Missing API token, required to retrieve Motive data", http.StatusBadRequest)
		return
	}

	// if we made it here, that means we got a valid token and we can run the pipeline
	for event, subscribers := range events {
		fmt.Printf("Processing motive data: %s, subscribers: %v\n", event, subscribers)
		for _, sub := range subscribers {
			fmt.Printf("event: %s, notifying subscriber: %s\n", event, sub)
		}

	}

	// return success msg to user
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := map[string]string{"status": "success", "msg": "Completed data load from Motive API for new customer"}
	json.NewEncoder(w).Encode(response)
}
