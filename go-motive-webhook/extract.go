package main

// setup motive client to make calls to motive API
import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

var BASE_URL = "https://api.gomotive.com/v1/"

// map of event names to endpoints and the parameters used in making requests to those endpoints
var ENDPOINTS = map[string]map[string]interface{}{
	"vehicles": {"endpoint": "vehicles", "params": map[string]string{}},
	"drivers":  {"endpoint": "users", "params": map[string]string{"role": "driver"}},
	"trailers": {"endpoint": "assets", "params": map[string]string{}},
}

func ExtractMotiveData(token string, event string) ([]interface{}, error) {

	// get endpoint and params from ENDPOINT map
	endpoint, params := getRequestDetails(event)
	log.Printf("Endpoint %s params %s", endpoint, params)

	resp, err := sendGETRequest(endpoint, params, token)
	log.Printf("Response status code: %d", resp.StatusCode)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	defer resp.Body.Close()

	// read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return nil, err
	}

	log.Println("Response Body: ", string(body))

	// the endpoint is the same name of the actual data key in the API response, so use it here to pull out the data
	result, _ := getDataFromResponse(endpoint, body)

	return result, nil
}

func getRequestDetails(event string) (string, map[string]string) {
	// get event request details from ENDPOINTS map
	if requestDetails, ok := ENDPOINTS[event]; ok {
		// Perform type assertions to get "endpoint" and "params"
		if endpoint, ok := requestDetails["endpoint"].(string); ok {
			params, _ := requestDetails["params"].(map[string]string)

			// encode params into expected format for GET request
			return endpoint, params
		}
	}
	return "", nil
}

func sendGETRequest(endpoint string, queryParams map[string]string, token string) (*http.Response, error) {
	// add endpoint to our BASEURL
	requestUrl := BASE_URL + endpoint

	// encode the query parameters
	values := url.Values{}
	for key, value := range queryParams {
		values.Add(key, value)
	}

	// create final url which includes encoded query params
	url := fmt.Sprintf("%s?%s", requestUrl, values.Encode())

	log.Printf("Motive API request URL: %s", url)

	// prepare get request with complete url
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// add auth token to the header of request
	authValue := "Bearer " + token
	request.Header.Add("Authorization", authValue)

	// Send the request
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func getDataFromResponse(key string, body []byte) ([]interface{}, error) {
	// unmarshal the JSON response into a slice of items
	var respData map[string]interface{}
	err := json.Unmarshal(body, &respData)
	if err != nil {
		return nil, err
	}

	// get actual data part from response
	data, exists := respData[key]
	if exists {
		log.Printf("Data for '%s' found in response", key)
	} else {
		msg := fmt.Sprintf("Key '%s' not found in the response\n", key)
		log.Println(msg)
		return nil, errors.New(msg)
	}

	// Check if the value is an array and get the number of items
	if finalData, ok := data.([]interface{}); ok {
		return finalData, nil
	}
	// if it's not an array, return an error
	msg := fmt.Sprintf("Value for key '%s' is not an array", key)
	return nil, errors.New(msg)
}
