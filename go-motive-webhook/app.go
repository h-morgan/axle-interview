package main

import (
	"fmt"
	"net/http"
)

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
	w.Write([]byte(`{"status": "ok", "msg": "we're up"}`))
}

func motivePipeline(w http.ResponseWriter, r *http.Request) {
	// Check if the request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
}
