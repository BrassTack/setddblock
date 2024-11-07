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

	// Test Lock and LastErr
	locker.Lock()
	require.Error(t, locker.LastErr(), "LastErr should return an error after failed lock")

	// Test ClearLastErr
	locker.ClearLastErr()
	require.NoError(t, locker.LastErr(), "LastErr should return nil after ClearLastErr")

	// Test Unlock and LastErr
	locker.Unlock()
	require.Error(t, locker.LastErr(), "LastErr should return an error after failed unlock")
}
}
