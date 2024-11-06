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
AWS_ACCESS_KEY_ID=dummy AWS_SECRET_ACCESS_KEY=dummy ./setddblock-macos-arm64 -nX --endpoint http://localhost:8000 ddb://test/lock_item_id /bin/sh -c 'echo "Lock acquired!"; sleep 30' & LOCK_PID=$!
echo "Initial lock acquired, sleeping for 30 seconds. PID: $LOCK_PID"

# Wait for a moment to ensure the lock is acquired
sleep 2

# Attempt to acquire the lock again to demonstrate it's locked
echo "Attempting to acquire lock again to demonstrate it's locked..."
if ! AWS_ACCESS_KEY_ID=dummy AWS_SECRET_ACCESS_KEY=dummy ./setddblock-macos-arm64 -nX --endpoint http://localhost:8000 ddb://test/lock_item_id /bin/sh -c 'echo "This should not run if lock is held"; exit 1'; then
  echo "Lock is held, as expected."
else
  echo "Error: Lock was acquired unexpectedly."
  exit 1
fi

# Simulate killing the process holding the lock
echo "Simulating process kill..."
kill $LOCK_PID

# Wait for a moment to ensure the lock is released
sleep 2

# Retry acquiring the lock until successful
echo "Retrying to acquire lock..."
while ! AWS_ACCESS_KEY_ID=dummy AWS_SECRET_ACCESS_KEY=dummy ./setddblock-macos-arm64 -nX --debug --timeout "100s" --endpoint http://localhost:8000 ddb://test/lock_item_id /bin/sh -c 'echo "Lock acquired after retry!"; exit 0'; do
  echo "Lock not acquired, retrying..."
  echo "Querying DynamoDB for lock item details..."
  AWS_ACCESS_KEY_ID=dummy AWS_SECRET_ACCESS_KEY=dummy aws dynamodb get-item --table-name test --key '{"ID": {"S": "lock_item_id"}}' --endpoint-url http://localhost:8000 --region ap-northeast-1
  sleep 1
done

# Stop DynamoDB Local
echo "Stopping DynamoDB Local..."
docker-compose down
