package logger

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	api "github.com/BryceDouglasJames/Cute-Logger/api"
	"github.com/stretchr/testify/require"
)

func TestNewLogAndNewSegment(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "log_test_dir")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a new Log instance with the temporary directory
	log, err := NewLog(tempDir)
	require.NoError(t, err)

	// Verify that a new segment is created if no log files exist
	require.Len(t, log.segmentList, 1, "Expected exactly one segment in the segment list")

	// Verify the active segment is correctly set
	require.Equal(t, log.segmentList[0], log.activeSegment, "Active segment should be the first segment in the list")

	// Create additional segments by simulating log files
	for i := 1; i <= 10; i++ {
		// Example offset values: 10, 20, 30, 40, ...
		offset := uint64(i * 10)

		// Create store files
		fileName := filepath.Join(tempDir, strconv.FormatUint(offset, 10)+".store")
		_, err = os.Create(fileName)
		require.NoError(t, err)

		// Create new segment from offset
		err = log.newSegment(offset)
		require.NoError(t, err)
	}

	// Verify the segments are correctly added
	require.Len(t, log.segmentList, 11, "Expected eleven segments in the segment list after adding more")

	// Verify that the offset of the active segment matches the expected value
	require.Equal(t, log.activeSegment.BaseOffset(), uint64(100), "Active segment should have the expected highest offset")

	// Verify the active segment is updated to the last created segment
	require.Equal(t, log.activeSegment, log.segmentList[len(log.segmentList)-1], "Active segment should be the last segment in the list")
}

func TestNewLogAppend(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "log_test_dir")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a new Log instance with the temporary directory
	log, err := NewLog(tempDir)
	require.NoError(t, err)

	// Verify that a new segment is created if no log files exist
	require.Len(t, log.segmentList, 1, "Expected exactly one segment in the segment list")

	// Verify the active segment is correctly set
	require.Equal(t, log.segmentList[0], log.activeSegment, "Active segment should be the first segment in the list")

	// Define a dummy record to append
	record := &api.Record{Value: []byte("dummy log entry")}

	prevActiveSegment := log.activeSegment

	// Append the record to the log
	off, err := log.Append(record)
	require.NoError(t, err)
	require.Equal(t, uint64(0), off)

	// Verify the active segment switches when full
	// Simulate appending records until the current active segment is full
	for !log.activeSegment.IsFull() {
		_, err := log.Append(record)
		require.NoError(t, err)
	}

	// Append another record to trigger a new segment creation
	_, err = log.Append(record)
	require.Error(t, err)

	// Verify a new segment was created and set as active
	require.NotEqual(t, prevActiveSegment, log.activeSegment, "A new segment should be active after the previous one is full")
	require.True(t, len(log.segmentList) > 1, "There should be more than one segment in the list after filling the first one")

}
