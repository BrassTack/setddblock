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

func getItemDetails(t *testing.T, client *dynamodb.Client, tableName, itemID string) (int64, string, error) {
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

// Function to acquire the initial lock in a forked process
func acquireInitialLock() {

	// Configure debug logging
	logger := log.New(os.Stdout, "[setddblock] ", log.LstdFlags|log.Lmsgprefix)
	filter := &logutils.LevelFilter{
	    Levels:   []logutils.LogLevel{"debug", "warn", "error"},
	    MinLevel: "debug",
	    Writer:   os.Stdout,
	}
	logger.SetOutput(filter)

	leaseDuration := 10 * time.Second
	locker, err := setddblock.New(
		"ddb://test/lock_item_id",
		setddblock.WithEndpoint("http://localhost:8000"),
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

	// Configure debug logging
	logger := log.New(os.Stdout, "[setddblock] ", log.LstdFlags|log.Lmsgprefix)
	filter := &logutils.LevelFilter{
	    Levels:   []logutils.LogLevel{"debug", "warn", "error"},
	    MinLevel: "debug",
	    Writer:   os.Stdout,
	}
	logger.SetOutput(filter)

	// Load AWS SDK DynamoDB client configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithEndpointResolver(
		aws.EndpointResolverFunc(func(service, region string) (aws.Endpoint, error) {
			if service == dynamodb.ServiceID {
				return aws.Endpoint{URL: "http://localhost:8000"}, nil
			}
			return aws.Endpoint{}, fmt.Errorf("unknown endpoint requested")
		}),
	))
	require.NoError(t, err, "Failed to load AWS SDK config")
	client := dynamodb.NewFromConfig(cfg)

	// Step 1: Check if we are in the main process or the forked process
	if os.Getenv("FORKED") == "1" {
		acquireInitialLock()
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
	initialTTL, initialRevision, err := getItemDetails(t, client, "test", "lock_item_id")
	require.NoError(t, err, "Failed to get item details")
	expireTime := time.Unix(initialTTL, 0)
	t.Logf("Initial DynamoDB item: REVISION=%s, TTL=%d (%s)", initialRevision, initialTTL, expireTime)

	retryInterval := 1 * time.Second
	retryCount := 0
	maxRetries := 100
	lockAcquired := false

	// Start retry loop
	for retryCount < maxRetries {
		retryCount++
		currentTime := time.Now()
		t.Logf("[Retry #%d] Attempting to acquire lock at %v (Unix: %d), expecting TTL expiration at %v (Unix: %d)",
			retryCount, currentTime, currentTime.Unix(), expireTime, initialTTL)

		locker, err := setddblock.New(
			"ddb://test/lock_item_id",
			setddblock.WithEndpoint("http://localhost:8000"),
			setddblock.WithLeaseDuration(5*time.Second),
			setddblock.WithDelay(false),
			setddblock.WithNoPanic(),
			// //simulate --debug flag
      setddblock.WithLogger(logger),

		)
		require.NoError(t, err, "Failed to create locker for retry")

		locker.Lock()
		if locker.LastErr() == nil {
			lockAcquired = true
			t.Logf("Lock acquired after TTL expiration on retry #%d", retryCount)
			break
		}

		// Check TTL to ensure it's stable and not being updated
		currentTTL, currentRevision, err := getItemDetails(t, client, "test", "lock_item_id")
		if err == nil {
			t.Logf("[Retry #%d] Current DynamoDB item: REVISION=%s, TTL=%d (%s)",
				retryCount, currentRevision, currentTTL, time.Unix(currentTTL, 0))
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
	require.GreaterOrEqual(t, actualAcquiredTime.Unix(), initialTTL, "Lock should only be acquired after TTL expiration")
}
