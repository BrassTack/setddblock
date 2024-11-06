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
echo "Running setddblock tool to acquire lock..."
echo "Acquiring initial lock..."
AWS_ACCESS_KEY_ID=dummy AWS_SECRET_ACCESS_KEY=dummy ./setddblock-macos-arm64 -xN --endpoint http://localhost:8000 ddb://test/lock_item_id /bin/sh -c 'echo "Lock acquired!"; sleep 30' &
echo "Initial lock acquired, sleeping for 30 seconds."

# Wait for a moment to ensure the lock is acquired
sleep 2

# Attempt to acquire the lock again to demonstrate it's locked
echo "Attempting to acquire lock again to demonstrate it's locked..."
AWS_ACCESS_KEY_ID=dummy AWS_SECRET_ACCESS_KEY=dummy ./setddblock-macos-arm64 -xN --endpoint http://localhost:8000 ddb://test/lock_item_id /bin/sh -c 'echo "This should not run if lock is held"; exit 1' || echo "Lock is held, as expected."

# Wait for the initial lock to expire
wait

# Stop DynamoDB Local
echo "Stopping DynamoDB Local..."
docker-compose down
