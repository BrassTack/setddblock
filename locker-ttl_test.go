// locker-ttl_test.go

package setddblock_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"
  "log"


	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/mashiike/setddblock"
	"github.com/stretchr/testify/require"
  "github.com/fujiwara/logutils"

)

/*
TestTTLExpirationLock aims to verify that a DynamoDB-based lock expires as expected based on its TTL.
The test follows these steps:
1. Acquire an initial lock with a defined TTL (5 seconds).
   - This lock is created using `DynamoDBLocker` and is intentionally left "unreleased" by killing the process to simulate a process crash.
2. Check the lock's `TTL` and `Revision` attributes directly in DynamoDB.
   - We use the AWS SDK to confirm the lock's TTL and verify that DynamoDB has recorded it.
3. Continuously attempt to acquire the same lock before the TTL expires.
   - Each acquisition attempt should fail until the TTL expires, confirming the lock is held until DynamoDB releases it.
4. Once the TTL expires, validate that the lock can now be reacquired.
   - This confirms that the time it took to reacquire the lock matches or exceeds the expected TTL, showing that the lock was released due to TTL expiration.
*/

func getItemDetails(client *dynamodb.Client, tableName, itemID string) (int64, string, error) {
	result, err := client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"ID": &types.AttributeValueMemberS{Value: itemID},
		},
	})
	if err != nil {
		return 0, "", fmt.Errorf("failed to query DynamoDB: %w", err)
	}

	var ttl int64
	if ttlAttr, ok := result.Item["ttl"].(*types.AttributeValueMemberN); ok {
		ttl, err = strconv.ParseInt(ttlAttr.Value, 10, 64)
	} else {
		return 0, "", fmt.Errorf("TTL attribute is missing or has an unexpected type")
	}

	revision := ""
	if revisionAttr, ok := result.Item["Revision"].(*types.AttributeValueMemberS); ok {
		revision = revisionAttr.Value
	} else {
		return 0, "", fmt.Errorf("Revision attribute is missing or has an unexpected type")
	}

	return ttl, revision, nil
}

func setupDynamoDBClient(t *testing.T) *dynamodb.Client {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithEndpointResolver(
		aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			if service == dynamodb.ServiceID {
				return aws.Endpoint{URL: dynamoDBURL}, nil
			}
			return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
		}),
	))
	require.NoError(t, err, "Failed to load AWS SDK config")
	return dynamodb.NewFromConfig(cfg)
}

func tryAcquireLock(t *testing.T, logger *log.Logger, retryCount int) bool {
	locker, err := setddblock.New(
		fmt.Sprintf("ddb://%s/%s", lockTableName, lockItemID),
		setddblock.WithEndpoint(dynamoDBURL),
		setddblock.WithLeaseDuration(5*time.Second),
		setddblock.WithDelay(false),
		setddblock.WithNoPanic(),
		setddblock.WithLogger(logger),
	)
	require.NoError(t, err, "Failed to create locker for retry")

	locker.Lock()
	if locker.LastErr() == nil {
		t.Logf("Lock acquired after TTL expiration on retry #%d", retryCount)
		return true
	}
	return false
}

const (
	leaseDuration   = 10 * time.Second
	retryInterval   = 1 * time.Second
	maxRetries      = 100
	dynamoDBURL     = "http://localhost:8000"
	lockItemID      = "lock_item_id"
	lockTableName   = "test"
)

func setupLogger(debug bool) *log.Logger {
	logger := log.New(os.Stdout, "[setddblock] ", log.LstdFlags)
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"debug", "warn", "error"},
		MinLevel: "warn",
		Writer:   os.Stdout,
	}
	if debug {
		filter.MinLevel = "debug"
	}
	logger.SetOutput(filter)
	return logger
}

func acquireInitialLock(logger *log.Logger) {
	locker, err := setddblock.New(
		fmt.Sprintf("ddb://%s/%s", lockTableName, lockItemID),
		setddblock.WithEndpoint(dynamoDBURL),
		setddblock.WithLeaseDuration(leaseDuration),
		setddblock.WithDelay(false),
		setddblock.WithNoPanic(),
		setddblock.WithLogger(logger),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create locker: %v\n", err)
		os.Exit(1)
	}
	locker.Lock()
	fmt.Println("Initial lock acquired; simulating lock hold indefinitely.")
	select {} // Keep the process alive to simulate a lock hold
}

// Test function with process forking and cleanup
func TestTTLExpirationLock(t *testing.T) {

	var retryCount int
	debug := false // Set this to false to disable --debug logging
	logger := setupLogger(debug)

	// Load AWS SDK DynamoDB client configuration
	client := setupDynamoDBClient(t)

	// Step 1: Check if we are in the main process or the forked process
	if os.Getenv("FORKED") == "1" {
		acquireInitialLock(logger)
		return
	}

	// Step 2: Fork the process to acquire and hold the initial lock
	t.Log("Forking process to acquire initial lock.")
	cmd := exec.Command(os.Args[0], "-test.run=TestTTLExpirationLock")
	cmd.Env = append(os.Environ(), "FORKED=1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Start(), "Failed to fork process for lock acquisition")

	// Allow the forked process time to acquire the lock
	t.Log("Waiting for forked process to acquire lock...")
	time.Sleep(3 * time.Second)

	// Step 3: Kill the forked process to simulate a crash
	t.Log("Killing forked process to simulate crash.")
	require.NoError(t, cmd.Process.Kill(), "Failed to kill forked process")

	// Confirm process termination
	processState, err := cmd.Process.Wait()
	if err != nil {
		t.Fatalf("Failed to confirm process termination: %v", err)
	}
	t.Logf("Forked process terminated with status: %v", processState)

	// Step 4: Log initial lock's TTL and revision from DynamoDB
	initialTTL, initialRevision, err := getItemDetails(client, lockTableName, lockItemID)
	require.NoError(t, err, "Failed to get item details")
	expireTime := time.Unix(initialTTL, 0)
	t.Logf("Initial DynamoDB item: REVISION=%s, TTL=%d (%s)", initialRevision, initialTTL, expireTime)

	lockAcquired := false

	// Start retry loop
	for retryCount < maxRetries {
		retryCount++
		currentTime := time.Now()
		t.Logf("[Retry #%d] Attempting lock acquisition at %v, expecting TTL expiration at %v",
			retryCount, currentTime.Format(time.RFC3339), expireTime.Format(time.RFC3339))

		lockAcquired = tryAcquireLock(t, logger, retryCount)
		if lockAcquired {
			break
		}

		// Check TTL to ensure it's stable and not being updated
		currentTTL, currentRevision, err := getItemDetails(client, lockTableName, lockItemID)
		if err == nil {
			t.Logf("[Retry #%d] Current item: REVISION=%s, TTL=%s",
				retryCount, currentRevision, time.Unix(currentTTL, 0).Format(time.RFC3339))
		} else {
			t.Logf("[Retry #%d] Failed to retrieve item details: %v", retryCount, err)
		}

		time.Sleep(retryInterval)
	}

	require.True(t, lockAcquired, "Expected to acquire lock after TTL expiration")
	actualAcquiredTime := time.Now()
	t.Logf("Lock finally acquired at %v (Unix: %d), expected TTL expiration at %v (Unix: %d)",
		actualAcquiredTime, actualAcquiredTime.Unix(), expireTime, initialTTL)

	// Log duration between TTL expiration and successful lock acquisition
	timeAfterTTL := actualAcquiredTime.Sub(expireTime)
	t.Logf("Time between TTL expiration and lock acquisition: %v", timeAfterTTL)
	require.LessOrEqual(t, timeAfterTTL.Seconds(), 3.0, "Time between TTL expiration and lock acquisition should not exceed 3 seconds")
	require.GreaterOrEqual(t, actualAcquiredTime.Unix(), initialTTL, "Lock should only be acquired after TTL expiration")
}
