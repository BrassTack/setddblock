package setddblock_test

import (
	"bytes"
	"log"
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
	options := createOptions(endpoint)
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
	logger := setupLogger()

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
		t.Logf("[%s] Function f1: Worker ID = %d has started processing", time.Now().Format(time.RFC3339), workerID)
		for i := 0; i < countMax; i++ {
			total1 += 1
			time.Sleep(10 * time.Millisecond)
		}
		lastTime1 = time.Now()
		t.Logf("[%s] Function f1: Worker ID = %d has finished processing", time.Now().Format(time.RFC3339), workerID)
	}
	f2 := func(workerID int, l sync.Locker) {
		defer func() {
			err := setddblock.Recover(recover())
			require.NoError(t, err)
		}()
		l.Lock()
		defer l.Unlock()
		t.Logf("[%s] Function f2: Worker ID = %d has started processing", time.Now().Format(time.RFC3339), workerID)

		for i := 0; i < countMax; i++ {
			total2 += 1
			time.Sleep(20 * time.Millisecond)
		}
		lastTime2 = time.Now()

		t.Logf("[%s] Function f2: Worker ID = %d has finished processing", time.Now().Format(time.RFC3339), workerID)
	}
	for i := 0; i < workerNum; i++ {
		wgEnd.Add(2)
		go func(workerID int) {
			defer wgEnd.Done()
			locker, err := setddblock.New(
				"ddb://test/item1",
				setddblock.WithDelay(true),
				setddblock.WithEndpoint(endpoint),
				options...,
			)
			require.NoError(t, err)
			wgStart.Wait()
			f1(workerID, locker)
		}(i + 1)
		go func(workerID int) {
			defer wgEnd.Done()
			locker, err := setddblock.New("ddb://test/item2", createOptions(endpoint)...)
			require.NoError(t, err)
			wgStart.Wait()
			f2(workerID, locker)
		}(i + 1)
	}
	wgStart.Done()
	wgEnd.Wait()
	t.Logf("[%s] %s", time.Now().Format(time.RFC3339), buf.String())
	require.EqualValues(t, workerNum*countMax, total1)
	require.EqualValues(t, workerNum*countMax, total2)
	t.Logf("[%s] Function f1: Last execution time = %d (%s)", time.Now().Format(time.RFC3339), lastTime1.Unix(), lastTime1.Format(time.RFC3339))
	t.Logf("[%s] Function f2: Last execution time = %d (%s)", time.Now().Format(time.RFC3339), lastTime2.Unix(), lastTime2.Format(time.RFC3339))
	require.True(t, lastTime1.After(lastTime2))
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

func TestNoPanic(t *testing.T) {
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
func createOptions(endpoint string) []func(*setddblock.Options) {
	options := []func(*setddblock.Options){
		setddblock.WithDelay(true),
		setddblock.WithEndpoint(endpoint),
		setddblock.WithLeaseDuration(500 * time.Millisecond),
	}
	// Check if logging is enabled and append the logger to options if true
	if enableLogging {
		logger := setupLogger()
		options = append(options, setddblock.WithLogger(logger))
	}
	return options
}
