package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
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

var ENV string = os.Getenv("ENV")

func main() {
	r := mux.NewRouter()
	// Define the routes
	r.HandleFunc("/", home).Methods(http.MethodGet)
	r.HandleFunc("/motive-pipeline", motivePipeline).Methods(http.MethodPost)

	// Start the server on port 8080
	log.Println("Server is listening on port 80...")
	log.Fatal(http.ListenAndServe(":80", r))
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

	// prepare response map
	var response map[string]string

	// if we made it here, that means we got a valid token and we can run the pipeline
	for event, subscribers := range events {
		log.Printf("Processing motive data: %s, subscribers: %v\n", event, subscribers)
		// retrieve data for specified event/resource
		data, extractErr := ExtractMotiveData(newToken.Value, event)
		if extractErr != nil {
			response = map[string]string{"status": "error", "msg": extractErr.Error()}
			break
		}
		// load the data to output storage and send to subscribers
		output, loadErr := LoadMotiveData(event, data)
		if loadErr != nil {
			response = map[string]string{"status": "error", "msg": loadErr.Error()}
			break
		}
		for _, sub := range subscribers {
			log.Printf("event: %s, notifying subscriber: %s\n", event, sub)
			err := RunWebhook(sub, event, output)
			if err != nil {
				response = map[string]string{"status": "error", "msg": err.Error()}
				break
			}
		}

	}

	// return success msg to user
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if len(response) == 0 {
		response = map[string]string{"status": "success", "msg": "Completed data load from Motive API for new customer"}
	}
	json.NewEncoder(w).Encode(response)
}
