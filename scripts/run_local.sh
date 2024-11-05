#!/usr/bin/env bash

set -e

# Start DynamoDB Local using Docker Compose
echo "Starting DynamoDB Local..."
docker-compose up -d ddb-local

# Wait for DynamoDB Local to be ready
echo "Waiting for DynamoDB Local to be ready..."
until curl -s http://localhost:8000; do
  sleep 1
done

echo "DynamoDB Local is ready."

# Run the setddblock tool against the local DynamoDB instance
echo "Running setddblock tool..."
AWS_ACCESS_KEY_ID=dummy AWS_SECRET_ACCESS_KEY=dummy ./setddblock-macos-arm64 -N --endpoint http://localhost:8000 ddb://test/lock_item_id echo "Hello, DynamoDB Local!"

# Stop DynamoDB Local
echo "Stopping DynamoDB Local..."
docker-compose down
