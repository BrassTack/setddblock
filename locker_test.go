package setddblock_test

import (
	"bytes"
	"log"
	"net/http"
	"os/exec"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fatih/color"
	"github.com/fujiwara/logutils"
	"github.com/mashiike/setddblock"
	"github.com/stretchr/testify/require"
)

func TestDDBLock(t *testing.T) {
	endpoint := checkDDBLocalEndpoint(t)
	defer func() {
		err := setddblock.Recover(recover())
		require.NoError(t, err)
	}()
	var buf bytes.Buffer
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"debug", "warn", "error"},
		MinLevel: "warn",
		ModifierFuncs: []logutils.ModifierFunc{
			logutils.Color(color.FgHiBlack),
			logutils.Color(color.FgYellow),
			logutils.Color(color.FgRed, color.Bold),
		},
		Writer: &buf,
	}
	logger := log.New(filter, "", log.LstdFlags|log.Lmsgprefix)

	var wgStart, wgEnd sync.WaitGroup
	wgStart.Add(1)
	var total1, total2 int
	var lastTime1, lastTime2 time.Time
	workerNum := 10
	countMax := 10
	f1 := func(workerID int, l sync.Locker) {
		defer func() {
			err := setddblock.Recover(recover())
			require.NoError(t, err)
		}()
		l.Lock()
		defer l.Unlock()
		t.Logf("f1 wroker_id = %d start", workerID)
		for i := 0; i < countMax; i++ {
			total1 += 1
			time.Sleep(10 * time.Millisecond)
		}
		lastTime1 = time.Now()
		t.Logf("f1 wroker_id = %d finish", workerID)
	}
	f2 := func(workerID int, l sync.Locker) {
		defer func() {
			err := setddblock.Recover(recover())
			require.NoError(t, err)
		}()
		l.Lock()
		defer l.Unlock()
		t.Logf("f2 wroker_id = %d start", workerID)

		for i := 0; i < countMax; i++ {
			total2 += 1
			time.Sleep(20 * time.Millisecond)
		}
		lastTime2 = time.Now()

		t.Logf("f2 wroker_id = %d finish", workerID)
	}
	for i := 0; i < workerNum; i++ {
		wgEnd.Add(2)
		go func(workerID int) {
			defer wgEnd.Done()
			locker, err := setddblock.New(
				"ddb://test/item1",
				setddblock.WithDelay(true),
				setddblock.WithEndpoint(endpoint),
				setddblock.WithLeaseDuration(500*time.Millisecond),
				setddblock.WithLogger(logger),
			)
			require.NoError(t, err)
			wgStart.Wait()
			f1(workerID, locker)
		}(i + 1)
		go func(workerID int) {
			defer wgEnd.Done()
			locker, err := setddblock.New(
				"ddb://test/item2",
				setddblock.WithDelay(true),
				setddblock.WithEndpoint(endpoint),
				setddblock.WithLeaseDuration(100*time.Millisecond),
				setddblock.WithLogger(logger),
			)
			require.NoError(t, err)
			wgStart.Wait()
			f2(workerID, locker)
		}(i + 1)
	}
	wgStart.Done()
	wgEnd.Wait()
	t.Log(buf.String())
	require.EqualValues(t, workerNum*countMax, total1)
	require.EqualValues(t, workerNum*countMax, total2)
	t.Logf("f1 last = %s", lastTime1)
	t.Logf("f2 last = %s", lastTime2)
	require.True(t, lastTime1.After(lastTime2))
	require.False(t, strings.Contains(buf.String(), "[error]"))
}

func TestUnlockStaleLockBasedOnTTL(t *testing.T) {
	endpoint := checkDDBLocalEndpoint(t)
	defer func() {
		err := setddblock.Recover(recover())
		require.NoError(t, err)
	}()
	var buf bytes.Buffer
	filter := &logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"debug", "warn", "error"},
		MinLevel: "warn",
		ModifierFuncs: []logutils.ModifierFunc{
			logutils.Color(color.FgHiBlack),
			logutils.Color(color.FgYellow),
			logutils.Color(color.FgRed, color.Bold),
		},
		Writer: &buf,
	}
	logger := log.New(filter, "", log.LstdFlags|log.Lmsgprefix)

	locker, err := setddblock.New(
		"ddb://test/stale_lock_item",
		setddblock.WithEndpoint(endpoint),
		setddblock.WithLeaseDuration(500*time.Millisecond),
		setddblock.WithLogger(logger),
	)
	require.NoError(t, err)

	// Acquire the lock
	lockGranted, err := locker.LockWithErr(context.Background())
	require.NoError(t, err)
	require.True(t, lockGranted)

	// Wait for the TTL to expire
	time.Sleep(1 * time.Second)

	// Attempt to unlock the stale lock
	err = locker.UnlockWithErr(context.Background())
	require.NoError(t, err)

	t.Log(buf.String())
	require.False(t, strings.Contains(buf.String(), "[error]"))
}
func checkDDBLocalEndpoint(t *testing.T) string {
	t.Helper()
	if endpoint := os.Getenv("DYNAMODB_LOCAL_ENDPOINT"); endpoint != "" {
		return endpoint
	}
	t.Log("ddb local endpoint not set. this test skip")
	t.SkipNow()
	return ""
}

func TestRunLocalScript(t *testing.T) {
	// Set up environment variables
	os.Setenv("AWS_ACCESS_KEY_ID", "dummy")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "dummy")
	os.Setenv("DYNAMODB_LOCAL_ENDPOINT", "http://localhost:8000")
	os.Setenv("AWS_DEFAULT_REGION", "ap-northeast-1")

	// Start DynamoDB Local using Docker Compose
	t.Log("Starting DynamoDB Local...")
	cmd := exec.Command("docker-compose", "up", "-d", "ddb-local")
	err := cmd.Run()
	require.NoError(t, err)

	// Wait for DynamoDB Local to be ready
	t.Log("Waiting for DynamoDB Local to be ready...")
	for {
		resp, err := http.Get("http://localhost:8000")
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		time.Sleep(1 * time.Second)
	}

	t.Log("DynamoDB Local is ready.")

	// Run the setddblock tool to acquire lock
	t.Log("Running setddblock tool to acquire lock...")
	cmd = exec.Command("./setddblock-macos-arm64", "-nX", "--endpoint", "http://localhost:8000", "ddb://test/lock_item_id", "/bin/sh", "-c", "echo 'Lock acquired!'; sleep 30")
	err = cmd.Start()
	require.NoError(t, err)
	lockPID := cmd.Process.Pid
	t.Logf("Initial lock acquired, sleeping for 30 seconds. PID: %d", lockPID)

	// Wait for a moment to ensure the lock is acquired
	time.Sleep(2 * time.Second)

	// Attempt to acquire the lock again to demonstrate it's locked
	t.Log("Attempting to acquire lock again to demonstrate it's locked...")
	cmd = exec.Command("./setddblock-macos-arm64", "-nX", "--endpoint", "http://localhost:8000", "ddb://test/lock_item_id", "/bin/sh", "-c", "echo 'This should not run if lock is held'; exit 1")
	err = cmd.Run()
	require.Error(t, err, "Lock is held, as expected.")

	// Simulate killing the process holding the lock
	t.Log("Simulating process kill...")
	err = exec.Command("kill", fmt.Sprintf("%d", lockPID)).Run()
	require.NoError(t, err)

	// Wait for a moment to ensure the lock is released
	time.Sleep(2 * time.Second)

	// Retry acquiring the lock until successful
	t.Log("Retrying to acquire lock...")
	retryCount := 0
	for {
		cmd = exec.Command("./setddblock-macos-arm64", "-nX", "--debug", "--endpoint", "http://localhost:8000", "ddb://test/lock_item_id", "/bin/sh", "-c", "echo 'Lock acquired after retry!'; exit 0")
		err = cmd.Run()
		if err == nil {
			break
		}
		retryCount++
		t.Logf("[retry %d] Lock not acquired, retrying...", retryCount)
		time.Sleep(1 * time.Second)
	}

	// Stop DynamoDB Local
	t.Log("Stopping DynamoDB Local...")
	cmd = exec.Command("docker-compose", "down")
	err = cmd.Run()
	require.NoError(t, err)
}
func TestInvalidEndpoint(t *testing.T) {
	defer func() {
		err := setddblock.Recover(recover())
		require.NoError(t, err, "check no panic")
	}()
	locker, err := setddblock.New(
		"ddb://test/item3",
		setddblock.WithEndpoint("http://localhost:12345"), //invalid remote endpoint
		setddblock.WithNoPanic(),
	)
	require.NoError(t, err)
	locker.Lock()
	require.Error(t, locker.LastErr())
	locker.ClearLastErr()
	locker.Unlock()
	require.Error(t, locker.LastErr())
}
}
