#!/bin/bash

# Exit on error and unset variables
set -eu

# Set up environment variables
export AWS_ACCESS_KEY_ID=dummy
export AWS_SECRET_ACCESS_KEY=dummy
export DYNAMODB_LOCAL_ENDPOINT=http://localhost:8000
export AWS_DEFAULT_REGION=ap-northeast-1

# Parse arguments for a specific test file
TEST_FILE="all"
while [[ $# -gt 0 ]]; do
  case "$1" in
    -t|--test)
      TEST_FILE="$2"
      shift 2
      ;;
    *)
      echo "Usage: $0 [-t|--test <test_file>]"
      exit 1
      ;;
  esac
done

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
if [[ "$TEST_FILE" == "all" ]]; then
  go test -v -race -timeout 30s ./...
else
  go test -v -race -timeout 30s "$TEST_FILE"
fi

# Stop DynamoDB Local
echo "Stopping DynamoDB Local..."
docker-compose down
