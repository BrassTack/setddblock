package setddblock_test

import (
	"context"
	"testing"
	"time"

	"github.com/mashiike/setddblock"
	"github.com/stretchr/testify/require"
)

func TestLockAcquisitionFailure(t *testing.T) {
	locker1, err := setddblock.New("ddb://test/item4", setddblock.WithLeaseDuration(1*time.Second), setddblock.WithNoPanic())
	require.NoError(t, err)
	locker2, err := setddblock.New("ddb://test/item4", setddblock.WithLeaseDuration(1*time.Second), setddblock.WithNoPanic())
	require.NoError(t, err)

	locker1.Lock()
	defer locker1.Unlock()

	lockGranted, err := locker2.LockWithErr(context.Background())
	require.Error(t, err)
	require.False(t, lockGranted)
}

func TestUnlockWithoutLock(t *testing.T) {
	locker, err := setddblock.New("ddb://test/item5", setddblock.WithLeaseDuration(1*time.Second))
	require.NoError(t, err)

	err = locker.UnlockWithErr(context.Background())
	require.Error(t, err)
}

func TestContextCancellation(t *testing.T) {
	locker, err := setddblock.New("ddb://test/item6", setddblock.WithLeaseDuration(1*time.Second))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	lockGranted, err := locker.LockWithErr(ctx)
	require.Error(t, err)
	require.False(t, lockGranted)
}

func TestLeaseDurationBoundaries(t *testing.T) {
	_, err := setddblock.New("ddb://test/item7", setddblock.WithLeaseDuration(50*time.Millisecond))
	require.Error(t, err)

	_, err = setddblock.New("ddb://test/item8", setddblock.WithLeaseDuration(15*time.Minute))
	require.Error(t, err)
}
