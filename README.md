# Axle Interview Project

This project is a backend webhook service that retrieves data from the [Motive API](https://developer.gomotive.com/reference/introduction), loads the data to blob storage, and notifies a callback URL of data loads.

The data we retrieve from the Motive API includes:

- vehicles
- drivers (from `/users` endpoint)
- trailers (from `/assets` endpoint)

## Table of Contents

- [1. Overview](#1-overview)
- [2. How to trigger the webhook](#2-how-to-trigger-the-webhook)
  - [2.1 Outputs of this service](#21-outputs-of-this-service)
- [3. How to run locally](#3-how-to-run-locally)
- [4. How to deploy to prod](#4-how-to-deploy-to-prod-aws-access-needed)
- [5. Service Docs](#5-serivce-docs)
  - [5.1 Endpoint: Home](#51-endpoint-home)
  - [5.2 Endpoint: motive-pipeline](#52-endpoint-motive-pipeline)
- [6. Future considerations](#6-future-considerations)

## 1. Overview

The project is deployed in AWS ECS, and data loads are stored in S3 blob storage. Data is also sent to the callback URL, with additional metadata about the load.

You'll notice there are 2 versions of the same service in this repo:

- Go version: [go-motive-webhook](https://github.com/h-morgan/axle-motive-webhook/tree/main/go-motive-webhook)
- Python version: [python-motive-webhook](https://github.com/h-morgan/axle-motive-webhook/tree/main/python-motive-webhook)

The reason for this is that Python is the language I'm most comfortable in, so I started by writing a first pass version in Python. For the purposes of the project, we're mainly interested in the Go version - this is the version currently running in AWS. Instructions in this main README will focus on the Go version, and instructions for the Python version will live in that [python-motive-webhook](https://github.com/h-morgan/axle-motive-webhook/tree/main/python-motive-webhook) directory.

## 2. How to trigger the webhook

To trigger the webhook, you should navigate in your browser to the following URL:

```
https://api.gomotive.com/oauth/authorize?client_id=98a670ed21a9b27a7e104160d61d51396577283d942b630202e12557a39a76f4&redirect_uri=https://eovvvgjxrp54hso.m.pipedream.net&response_type=code&scope=users.read%20vehicles.read%20assets.read
```

This URL makes a request to start the Oauth2.0 process with Motive. It contains:

- a `client_id`, identifying our application to Motive
- a `redirect_uri`, which tells Motive where to go after the user authenticates
- scopes required for our application to fetch the data required (users, vehicles, and assets)

The URL should bring you to a Motive login screen, where you should provide your username and password. Upon successful authentication, an access code is then automatically sent to the `redirect_uri` provided, which is a pipedream trigger we've configured to handle incoming authentication redirects.

If authentication is successful, an access token will be retrieved via a step in the Pipedream pipeline. The pipeline then sends the token in a POST request to our application running in AWS.

### 2.1 Outputs of this service

This service does 2 things:

1. Updates a callback URL for each "event" or data extraction/load process (vehicles, drivers, trailers). The callback URL is sent the data itself, and some metadata about the load
2. Saves the retrieved data to s3 blob storage

The data is returned to the callback URL and stored in S3 in the same JSON payload format:

```json
{
  "resource": "resource_name",
  "processed_at": "date/time of processing",
  "num_items": "num items retrieved from API for resource",
  "data": { "dict": "containing api data response for resource" }
}
```

## 3. How to run locally

Note: you can't use the Pipedream pipeline when running locally. You need to first run Oauth process (either in Postman, or by triggering the Pipedream pipeline). Use the retrieved access token to make requests in Postman/Insomnia or via cURL.

To run the go webhook service locally, clone this repo and build the docker image:

```
docker build --file go.Dockerfile --tag axle .
```

Then run the docker container from the image:

```
docker run -p 80:80 --name axle axle
```

This spins up the Go service on your localhost, where it will wait for incoming requests with new tokens.

## 4. How to deploy to prod (AWS access needed)

Prerequisites:

- Access to AWS account where this service is running
- aws cli configured
- Docker

Once all pre-reqs are met, run the build script:

```
./build-push-go.sh
```

This uploads a new version of the image with the `:latest` tag. You then need to manually kill the running version of the ECS service. A new version of the ECS service will automatically spin up using the latest image.

## 5. Service Docs

This section contains documentation on the endpoints for this service, and their expected inputs/outputs.

### 5.1 Endpoint: Home

Local URL: 127.0.0.1
AWS URL: http://ec2-3-144-118-104.us-east-2.compute.amazonaws.com

Should receive the following response:

```json
{
  "msg": "we're up",
  "status": "success"
}
```

### 5.2 Endpoint: motive-pipeline

To run the webhook/pipeline, make a request to:

Local URL: 127.0.0.1/motive-pipeline
AWS URL: http://ec2-3-144-118-104.us-east-2.compute.amazonaws.com/motive-pipeline

JSON body:

```json
{
  "token": "{{ token }}"
}
```

Token value should be the access token retrieved from the oauth process.

A successful run should result in data being sent to the callback URL, files saved to S3, and the following response:

```json
{
  "msg": "Completed data load from Motive API for new customer",
  "status": "success"
}
```

If an error is recieved during this process, you will receive an error response with a message containing some hopefully descriptive info about the error.

## 6. Future considerations

In order to get this project running in a timely manner, I cut some corners I normally would not have cut. This was also my first time developing a service in Go, so there was a learning curve there that took some time away from time I would have otherwise allocated to further development.

Some of the things I would like to do to improve/expand this service if I was spending more time on it:

- Unit tests. I typically use pytest for unit testing (most recent example of this in my [hntpy](https://github.com/h-morgan/hntpy/tree/main/tests) project) but cut this corner in Go for now since I'm not familiar with best practices there.
- Logic to handle access token refreshing. I left this out because of the small amount of data the service needed to handle, and the expiration time of the tokens being plenty long enough to handle the requests. Production would definitely need this as you can't know upfront if a customer might have large amounts of data that take longer than token expiration time to process. This would require expanding the `/motive-pipeline` expected POST request body to include additional params like `refresh_token` and `expires_at` time.
- Pagination handling on API requests to Motive. This was something I implemented in the python version of the app, but cut this corner when developing the Go app due to the small amount of data I needed to process not requiring it. Production would definitely need this though, as you never know how much data a customer may have.
- Better error handling
- A way to identify the customer so their data can be filed and retrieved. If this is run for mutliple customers/Motive accounts it's not clear who's data we're retrieving.
