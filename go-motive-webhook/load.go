package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var S3_BUCKET string = "axle-motive-data"

func LoadMotiveData(event string, data []interface{}) ([]byte, error) {
	var loadInfo = make(map[string]interface{})
	loadInfo["resource"] = event
	loadInfo["data"] = data

	// get current time to record processed_at
	processedAt := time.Now()
	loadInfo["processed_at"] = processedAt

	// get number of items, add to load info response
	loadInfo["num_items"] = len(data)

	// save data to s3 (skips if we're not running in prod)
	key := "data/" + event + ".json"
	jsonData, s3Err := saveToS3(S3_BUCKET, key, loadInfo)
	if s3Err != nil {
		fmt.Println("Error:", s3Err)
		return nil, s3Err
	}

	return jsonData, nil
}

func RunWebhook(url string, event string, jsonData []byte) error {

	// submit post request to subscriber URL
	response, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}
	defer response.Body.Close()

	// read response
	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return err
	}

	fmt.Println("Status Code:", response.Status)
	fmt.Println("Response Body:", string(responseData))
	return nil
}

func saveToS3(bucket string, s3Key string, data map[string]interface{}) ([]byte, error) {

	// convert final response to JSON payload
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println("Error:", err)
		return nil, err
	}

	// only upload to s3 if we're not running in dev
	if ENV == "PROD" {
		// if we're running in prod, load data to s3
		log.Println("Saving files to AWS S3")
		// create aws session
		s, err := session.NewSession(&aws.Config{
			Region: aws.String("us-east-2"),
		})
		if err != nil {
			log.Println("Error creating session in AWS, cannot store files to s3.")
			log.Println(err)
			return nil, err
		}

		// create an S3 service client
		s3Client := s3.New(s)

		// upload the JSON data to S3
		_, err = s3Client.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(s3Key),
			Body:   bytes.NewReader(jsonData),
		})
		if err != nil {
			log.Println("Error storing files to s3.")
			log.Println(err)
			return nil, err
		}

		fmt.Printf("Map stored in S3 at s3://%s/%s\n", bucket, s3Key)
	}
	return jsonData, nil
}
