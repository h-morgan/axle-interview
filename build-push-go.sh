#!/usr/bin/env bash
# script to build image and push to ecr repo

echo "Building axle-webhook service, version: go"

ECR_HOST=105864845804.dkr.ecr.us-east-2.amazonaws.com
IMAGE_NAME=axle-webhook-go
REPO="$ECR_HOST/$IMAGE_NAME:latest"

echo "Building and pushing to AWS ECR: $REPO"

# build

# generate new aws password token - needed for pushing to our aws acct
aws ecr get-login-password --region us-east-2 | docker login --username AWS --password-stdin $ECR_HOST

# build image - adaptation on docker build cmd to specify linux amd64 build (override mac m1 build issues)
docker buildx build --platform=linux/amd64 -f go.Dockerfile -t $IMAGE_NAME .

docker tag $IMAGE_NAME:latest $ECR_HOST/$IMAGE_NAME:latest


# push up to aws ecr
docker push $REPO
