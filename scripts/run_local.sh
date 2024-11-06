#!/usr/bin/env bash

set -e

start_time=$SECONDS
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

# Function to get item details from DynamoDB
get_item_details() {
  item_details=$(AWS_ACCESS_KEY_ID=dummy AWS_SECRET_ACCESS_KEY=dummy aws dynamodb get-item --table-name test --key '{"ID": {"S": "lock_item_id"}}' --endpoint-url http://localhost:8000 --region ap-northeast-1 --output text)
  if [ -z "$item_details" ]; then
    echo "No item found in DynamoDB for lock_item_id"
    return 1
  fi
  revision=$(echo "$item_details" | grep "REVISION" | awk '{print $2}')
  ttl=$(echo "$item_details" | grep "TTL" | awk '{print $2}')
  echo "REVISION: $revision, TTL: $ttl (Unix timestamp) at $(date +%s) expires $(date -r $ttl)"
}
echo "Checking DynamoDB item details before retrying to acquire the lock..."
echo "Checking DynamoDB item details before retrying to acquire the lock..."
get_item_details || echo "Skipping item details check due to missing record."

# Retry acquiring the lock until successful
echo "Retrying to acquire lock..."
retry_count=0
start_time=$(date +%s)
while ! AWS_ACCESS_KEY_ID=dummy AWS_SECRET_ACCESS_KEY=dummy ./setddblock-macos-arm64 -nX --debug --timeout "100s" --endpoint http://localhost:8000 ddb://test/lock_item_id /bin/sh -c 'echo "Lock acquired after retry!"; exit 0'; do
  retry_count=$((retry_count + 1))
  elapsed_time=$SECONDS
  echo "[$(date +%Y-%m-%dT%H:%M:%S)] [retry $retry_count][${elapsed_time}s] Lock not acquired, retrying..."
  echo "[$(date +%Y-%m-%dT%H:%M:%S)] [retry $retry_count][${elapsed_time}s] Querying DynamoDB for lock item details (table: test, item ID: lock_item_id)..."
  get_item_details || echo "Skipping item details check due to missing record."
  sleep 1
done

# Check DynamoDB item details after retrying to acquire the lock
echo "Checking DynamoDB item details after retrying to acquire the lock..."
echo "Checking DynamoDB item details after retrying to acquire the lock..."
get_item_details || echo "Skipping item details check due to missing record."

# Stop DynamoDB Local
elapsed_time=$SECONDS
echo "[$(date +%Y-%m-%dT%H:%M:%S) [${elapsed_time}s]] Stopping DynamoDB Local..."
docker-compose down
