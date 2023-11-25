package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// this will be the load metadata that gets built and sent to subscribers

func LoadMotiveData(event string, data []interface{}) (map[string]interface{}, error) {
	var loadInfo = make(map[string]interface{})
	loadInfo["resource"] = event
	loadInfo["data"] = data

	// get current time to record processed_at
	processedAt := time.Now()
	loadInfo["processed_at"] = processedAt

	// get number of items, add to load info response
	loadInfo["num_items"] = len(data)

	// TODO: load data into S3, persistant storage, etc.
	return loadInfo, nil
}

func RunWebhook(url string, data map[string]interface{}) {

	// convert final response to JSON payload
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println("Error:", err)
		return
	}
	// submit post request to subscriber URL
	response, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer response.Body.Close()

	// read response
	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	fmt.Println("Status Code:", response.Status)
	fmt.Println("Response Body:", string(responseData))
}
