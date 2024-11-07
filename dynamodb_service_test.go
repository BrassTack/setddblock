package setddblock_test

import (
	"context"
	"testing"
	"time"

	"github.com/mashiike/setddblock"
	"github.com/stretchr/testify/require"
)

func TestDynamoDBService_CreateLockTable(t *testing.T) {
	opts := setddblock.WithEndpoint("http://localhost:8000")
	svc, err := setddblock.NewDynamoDBService(opts)
	require.NoError(t, err)

	err = svc.CreateLockTable(context.TODO(), "test-table")
	require.NoError(t, err, "CreateLockTable should not return an error")
}

func TestDynamoDBService_WaitLockTableActive(t *testing.T) {
	opts := setddblock.WithEndpoint("http://localhost:8000")
	svc, err := setddblock.NewDynamoDBService(opts)
	require.NoError(t, err)

	err = svc.WaitLockTableActive(context.TODO(), "test-table")
	require.NoError(t, err, "WaitLockTableActive should not return an error")
}

func TestDynamoDBService_AquireLock(t *testing.T) {
	opts := setddblock.WithEndpoint("http://localhost:8000")
	svc, err := setddblock.NewDynamoDBService(opts)
	require.NoError(t, err)

	input := &setddblock.LockInput{
		TableName:     "test-table",
		ItemID:        "test-item",
		LeaseDuration: 5 * time.Second,
		Revision:      "test-revision",
	}

	output, err := svc.AquireLock(context.TODO(), input)
	require.NoError(t, err, "AquireLock should not return an error")
	require.NotNil(t, output, "AquireLock should return a valid output")
}

func TestDynamoDBService_SendHeartbeat(t *testing.T) {
	opts := setddblock.WithEndpoint("http://localhost:8000")
	svc, err := setddblock.NewDynamoDBService(opts)
	require.NoError(t, err)

	input := &setddblock.LockInput{
		TableName:     "test-table",
		ItemID:        "test-item",
		LeaseDuration: 5 * time.Second,
		Revision:      "test-revision",
		PrevRevision:  aws.String("prev-revision"),
	}

	output, err := svc.SendHeartbeat(context.TODO(), input)
	require.NoError(t, err, "SendHeartbeat should not return an error")
	require.NotNil(t, output, "SendHeartbeat should return a valid output")
}

func TestDynamoDBService_ReleaseLock(t *testing.T) {
	opts := setddblock.WithEndpoint("http://localhost:8000")
	svc, err := setddblock.NewDynamoDBService(opts)
	require.NoError(t, err)

	input := &setddblock.LockInput{
		TableName:    "test-table",
		ItemID:       "test-item",
		PrevRevision: aws.String("prev-revision"),
	}

	err = svc.ReleaseLock(context.TODO(), input)
	require.NoError(t, err, "ReleaseLock should not return an error")
}
