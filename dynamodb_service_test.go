package setddblock_test

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/mashiike/setddblock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockDynamoDBClient is a mock implementation of the DynamoDB client
type MockDynamoDBClient struct {
	mock.Mock
}

func (m *MockDynamoDBClient) DescribeTable(ctx context.Context, input *dynamodb.DescribeTableInput, opts ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*dynamodb.DescribeTableOutput), args.Error(1)
}

func (m *MockDynamoDBClient) CreateTable(ctx context.Context, input *dynamodb.CreateTableInput, opts ...func(*dynamodb.Options)) (*dynamodb.CreateTableOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*dynamodb.CreateTableOutput), args.Error(1)
}

func (m *MockDynamoDBClient) PutItem(ctx context.Context, input *dynamodb.PutItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*dynamodb.PutItemOutput), args.Error(1)
}

func (m *MockDynamoDBClient) UpdateItem(ctx context.Context, input *dynamodb.UpdateItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*dynamodb.UpdateItemOutput), args.Error(1)
}

func (m *MockDynamoDBClient) DeleteItem(ctx context.Context, input *dynamodb.DeleteItemInput, opts ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*dynamodb.DeleteItemOutput), args.Error(1)
}

func TestDynamoDBService_CreateLockTable(t *testing.T) {
	mockClient := new(MockDynamoDBClient)
	svc := &setddblock.DynamoDBService{Client: mockClient}

	mockClient.On("CreateTable", mock.Anything, mock.Anything).Return(&dynamodb.CreateTableOutput{}, nil)

	err := svc.CreateLockTable(context.TODO(), "test-table")
	require.NoError(t, err, "CreateLockTable should not return an error")
	mockClient.AssertExpectations(t)
}

func TestDynamoDBService_WaitLockTableActive(t *testing.T) {
	mockClient := new(MockDynamoDBClient)
	svc := &setddblock.DynamoDBService{Client: mockClient}

	mockClient.On("DescribeTable", mock.Anything, mock.Anything).Return(&dynamodb.DescribeTableOutput{
		Table: &types.TableDescription{
			TableStatus: types.TableStatusActive,
		},
	}, nil)

	err := svc.WaitLockTableActive(context.TODO(), "test-table")
	require.NoError(t, err, "WaitLockTableActive should not return an error")
	mockClient.AssertExpectations(t)
}

func TestDynamoDBService_AquireLock(t *testing.T) {
	mockClient := new(MockDynamoDBClient)
	svc := &setddblock.DynamoDBService{Client: mockClient}

	mockClient.On("PutItem", mock.Anything, mock.Anything).Return(&dynamodb.PutItemOutput{}, nil)

	lockInput := &setddblock.LockInput{
		TableName:     "test-table",
		ItemID:        "test-item",
		LeaseDuration: 5 * time.Second,
		Revision:      "test-revision",
	}

	lockOutput, err := svc.AquireLock(context.TODO(), lockInput)
	require.NoError(t, err, "AquireLock should not return an error")
	require.True(t, lockOutput.LockGranted, "Lock should be granted")
	mockClient.AssertExpectations(t)
}

func TestDynamoDBService_SendHeartbeat(t *testing.T) {
	mockClient := new(MockDynamoDBClient)
	svc := &setddblock.DynamoDBService{Client: mockClient}

	mockClient.On("UpdateItem", mock.Anything, mock.Anything).Return(&dynamodb.UpdateItemOutput{}, nil)

	lockInput := &setddblock.LockInput{
		TableName:     "test-table",
		ItemID:        "test-item",
		LeaseDuration: 5 * time.Second,
		Revision:      "test-revision",
		PrevRevision:  aws.String("prev-revision"),
	}

	lockOutput, err := svc.SendHeartbeat(context.TODO(), lockInput)
	require.NoError(t, err, "SendHeartbeat should not return an error")
	require.True(t, lockOutput.LockGranted, "Heartbeat should be sent successfully")
	mockClient.AssertExpectations(t)
}

func TestDynamoDBService_ReleaseLock(t *testing.T) {
	mockClient := new(MockDynamoDBClient)
	svc := &setddblock.DynamoDBService{Client: mockClient}

	mockClient.On("DeleteItem", mock.Anything, mock.Anything).Return(&dynamodb.DeleteItemOutput{}, nil)

	lockInput := &setddblock.LockInput{
		TableName:    "test-table",
		ItemID:       "test-item",
		PrevRevision: aws.String("prev-revision"),
	}

	err := svc.ReleaseLock(context.TODO(), lockInput)
	require.NoError(t, err, "ReleaseLock should not return an error")
	mockClient.AssertExpectations(t)
}
