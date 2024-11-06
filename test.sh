#!/bin/bash

# Set up environment variables
export AWS_ACCESS_KEY_ID=dummy
export AWS_SECRET_ACCESS_KEY=dummy
export DYNAMODB_LOCAL_ENDPOINT=http://localhost:8000
export AWS_DEFAULT_REGION=ap-northeast-1

# Start DynamoDB Local using Docker Compose
echo "Starting DynamoDB Local..."
docker-compose up -d ddb-local

# Wait for DynamoDB Local to be ready
echo "Waiting for DynamoDB Local to be ready..."
until curl -s http://localhost:8000; do
  sleep 1
done

echo "DynamoDB Local is ready."

# Run tests
echo "Running tests..."
go test -race -timeout 30s ./...

# Stop DynamoDB Local
echo "Stopping DynamoDB Local..."
docker-compose down
