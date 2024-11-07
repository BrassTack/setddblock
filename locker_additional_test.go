package setddblock_test

import (
	"context"
	"testing"
	"time"

	"github.com/mashiike/setddblock"
	"github.com/stretchr/testify/require"
)

func TestGenerateRevision(t *testing.T) {
	locker, err := setddblock.New("ddb://test/item", setddblock.WithNoPanic())
	require.NoError(t, err)

	rev, err := locker.GenerateRevision()
	require.NoError(t, err, "GenerateRevision should not return an error")
	require.NotEmpty(t, rev, "GenerateRevision should return a non-empty string")
}

func TestLastErrAndClearLastErr(t *testing.T) {
	locker, err := setddblock.New("ddb://test/item", setddblock.WithNoPanic())
	require.NoError(t, err)

	locker.Lock()
	require.Error(t, locker.LastErr(), "LastErr should return an error after failed lock")

	locker.ClearLastErr()
	require.NoError(t, locker.LastErr(), "LastErr should return nil after ClearLastErr")
}

func TestBailout(t *testing.T) {
	locker, err := setddblock.New("ddb://test/item", setddblock.WithNoPanic())
	require.NoError(t, err)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("bailout should not panic when NoPanic is set")
		}
	}()

	locker.Bailout(errors.New("test error"))
	require.Error(t, locker.LastErr(), "LastErr should return the error set by bailout")
}
