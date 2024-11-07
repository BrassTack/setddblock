package setddblock_test

import (
	"context"
	"testing"
	"time"

	"github.com/mashiike/setddblock"

	"github.com/stretchr/testify/require"
)

func TestGenerateRevision(t *testing.T) {

	locker, err := setddblock.New(
		"ddb://test/item",
		setddblock.WithNoPanic(),
	)
	require.NoError(t, err)

	revision, err := locker.GenerateRevision()
	require.NoError(t, err)
	require.NotEmpty(t, revision, "Generated revision should not be empty")
	locker, err := setddblock.New(
		"ddb://test/item",
		setddblock.WithNoPanic(),
	)
	require.NoError(t, err)

	locker.Lock()
	require.Error(t, locker.LastErr(), "LastErr should return an error after failed lock")

	locker.ClearLastErr()
	require.NoError(t, locker.LastErr(), "LastErr should return nil after ClearLastErr")
}

func TestLastErrAndClearLastErr(t *testing.T) {
}

func TestBailout(t *testing.T) {

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("bailout should not panic when NoPanic is set")
		}
	}()

	locker, err := setddblock.New(
		"ddb://test/item",
		setddblock.WithNoPanic(),
	)
	require.NoError(t, err)

	locker.Lock()
	require.Error(t, locker.LastErr(), "LastErr should return an error after failed lock")

	locker.ClearLastErr()
	require.NoError(t, locker.LastErr(), "LastErr should return nil after ClearLastErr")

	locker.Unlock()
	require.Error(t, locker.LastErr(), "LastErr should return an error after failed unlock")
}
