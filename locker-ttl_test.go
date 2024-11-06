// locker-ttl_test.go

package setddblock_test

import (
	"testing"
	"time"

	"github.com/mashiike/setddblock"
	"github.com/stretchr/testify/require"
)

func TestTTLExpirationLock(t *testing.T) {
	// Step 1: Acquire initial lock with a short TTL
	locker, err := setddblock.New(
		"ddb://test/lock_item_id",
		setddblock.WithEndpoint("http://localhost:8000"),
		setddblock.WithLeaseDuration(5*time.Second), // TTL of 5 seconds
	)
	require.NoError(t, err, "Failed to create locker")
	require.NoError(t, locker.Lock(), "Failed to acquire initial lock")
	t.Log("Initial lock acquired. Simulating process termination by not unlocking.")

	// Capture the expected expiration time based on the TTL
	expectedExpiration := time.Now().Add(5 * time.Second)
	t.Logf("Expected TTL expiration time: %v", expectedExpiration)

	// Step 2: Aggressively attempt to acquire the lock before TTL expiration
	lockAcquired := false
	for time.Now().Before(expectedExpiration.Add(2 * time.Second)) { // Buffer after TTL

		// Attempt to acquire a new lock with the same lock ID
		lockAttempt, err := setddblock.New(
			"ddb://test/lock_item_id",
			setddblock.WithEndpoint("http://localhost:8000"),
		)
		require.NoError(t, err, "Failed to create locker for retry")

		// Try to lock again
		if err := lockAttempt.Lock(); err == nil {
			// Successfully acquired lock, indicating TTL expiration
			t.Logf("Lock successfully acquired after TTL expired at %v", time.Now())
			lockAttempt.Unlock() // Clean up
			lockAcquired = true
			break
		} else {
			t.Logf("Lock still held at %v; retrying...", time.Now())
		}
		time.Sleep(200 * time.Millisecond) // Short wait before retrying
	}

	// Step 3: Confirm that the lock was acquired after the expected TTL expiration
	require.True(t, lockAcquired, "Expected to acquire lock after TTL expiration")
	finalTime := time.Now()
	t.Logf("Lock acquisition confirmed at %v (expected after %v)", finalTime, expectedExpiration)
	require.GreaterOrEqual(t, finalTime.Unix(), expectedExpiration.Unix(), "Lock acquired too early, before TTL expired")
}
