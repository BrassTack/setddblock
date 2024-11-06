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
echo "Running tests for the setddblock package..."
go test -v -race -timeout 30s ./... | tee test_output.log

echo "Tests completed. Summary:"
grep -E "^(ok|FAIL|=== RUN|--- PASS|--- FAIL)" test_output.log

# Stop DynamoDB Local
echo "Stopping DynamoDB Local..."
docker-compose down
