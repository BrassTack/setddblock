package setddblock_test

import (
	"testing"

	"github.com/mashiike/setddblock"

	"github.com/stretchr/testify/require"
)

func TestLockerFunctions(t *testing.T) {
	locker, err := setddblock.New(
		"ddb://test/item",
		setddblock.WithNoPanic(),
	)
	require.NoError(t, err)

	// Test successful lock acquisition
	locker.Lock()
	require.NoError(t, locker.LastErr(), "Lock should be acquired without error")

	// Test lock release
	locker.Unlock()
	require.NoError(t, locker.LastErr(), "Unlock should be successful without error")

	// Test re-acquisition of lock
	locker.Lock()
	require.NoError(t, locker.LastErr(), "Re-acquisition of lock should be successful")

	// Test error handling by simulating a failure
	locker.Unlock()
	locker.Lock()
	require.Error(t, locker.LastErr(), "Simulated error should be captured by LastErr")

	// Test concurrency by attempting to lock from multiple goroutines
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			locker.Lock()
			defer locker.Unlock()
			require.NoError(t, locker.LastErr(), "Concurrent lock should be acquired without error")
		}()
	}
	wg.Wait()
}
