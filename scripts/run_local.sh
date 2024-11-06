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
AWS_ACCESS_KEY_ID=dummy AWS_SECRET_ACCESS_KEY=dummy ./setddblock-macos-arm64 --debug -xN --endpoint http://localhost:8000 ddb://test/lock_item_id /bin/sh -c 'echo "Hello, DynamoDB Local!";sleep 300' &
echo ran now sleep 5
sleep 5
echo try again
AWS_ACCESS_KEY_ID=dummy AWS_SECRET_ACCESS_KEY=dummy ./setddblock-macos-arm64 --debug -xN --endpoint http://localhost:8000 ddb://test/lock_item_id /bin/sh -c 'echo "Hello, DynamoDB Local 22222!";sleep 300'
echo ran 2



wait
wait




# Stop DynamoDB Local
echo "Stopping DynamoDB Local..."
docker-compose down
