package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	api "github.com/BryceDouglasJames/Cute-Logger/api"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
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
	tempDir, err := os.MkdirTemp("", "log_test_again")
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

	// Verify the record is correctly appended to the active segment
	readBack, err := log.activeSegment.Read(off)
	require.NoError(t, err)
	require.Equal(t, record.Value, readBack.Value, "The read-back record should match the appended record")

	// Verify the active segment switches when full
	// Simulate appending records until the current active segment is full
	for !log.segmentList[0].IsFull() {
		_, err := log.Append(record)
		require.NoError(t, err)
	}

	// Append another record to trigger a new segment creation
	_, err = log.Append(record)
	require.NoError(t, err)

	// Verify a new segment was created and set as active
	require.NotEqual(t, prevActiveSegment, log.activeSegment, "A new segment should be active after the previous one is full")
	require.True(t, len(log.segmentList) > 1, "There should be more than one segment in the list after filling the first one")

}

func TestLogRead(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "log_test_dir")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a new log instance with the temporary directory
	log, err := NewLog(tempDir)
	require.NoError(t, err)

	// Append a record to ensure there is at least one segment
	initialRecord := &api.Record{Value: []byte("initial record")}
	offset, err := log.Append(initialRecord)
	require.NoError(t, err)

	// Attempt to read back the record based on its offset
	readRecord, err := log.Read(offset)
	require.NoError(t, err)

	// Verify that the read record matches the initial record
	require.Equal(t, initialRecord.Value, readRecord.Value, "The read record should match the initial record")
}

func TestLogClose(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "log_test_dir")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a new log instance with the temporary directory
	log, err := NewLog(tempDir)
	require.NoError(t, err)

	// Attempt to close without errors
	require.NoError(t, log.Close(), "closing log should not produce an error")
}

func TestLogDelete(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "log_test_dir")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a new log instance with the temporary directory
	log, err := NewLog(tempDir)
	require.NoError(t, err)

	// Attempt to delete logger and make sure directory no longer exists
	require.NoError(t, log.Delete(), "deleting log should not produce an error")
	_, err = os.Stat(tempDir)
	require.True(t, os.IsNotExist(err), "log directory should be removed after delete")
}

func TestLogReset(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "log_test_dir")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a new log instance with the temporary directory
	log, err := NewLog(tempDir)
	require.NoError(t, err)

	// Simulate adding data to the log
	dummyRecord := &api.Record{Value: []byte("test")}
	_, err = log.Append(dummyRecord)
	require.NoError(t, err)

	// Reset the logger
	require.NoError(t, log.Reset(), "resetting log should not produce an error")

	// Check the segment list contains only one segment.
	require.Len(t, log.segmentList, 1, "Expected exactly one segment after reset")
}

func TestLogTruncate(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "log_truncate_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Initialize the log with the temporary directory
	log, err := NewLog(tempDir)
	require.NoError(t, err)

	// Append some records to generate segments
	for i := 0; i < 5; i++ {
		_, err := log.Append(&api.Record{Value: []byte(fmt.Sprintf("record %d", i))})
		require.NoError(t, err)
	}

	// Simulate truncating the log to remove early segments
	err = log.Truncate(2)
	require.NoError(t, err)

	// Verify that segments with nextOffset <= 3 are removed
	for _, s := range log.segmentList {
		require.True(t, s.NextOffset() > 3, "Segment with nextOffset <= 3 should have been truncated")
	}

}

func TestLogReader(t *testing.T) {
	// Create a temporary directory for the log
	tempDir, err := os.MkdirTemp("", "log_test_reader")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a new log instance.
	log, err := NewLog(tempDir)
	require.NoError(t, err)

	// Create a record to append to the log
	append := &api.Record{
		Value: []byte("hello world"),
	}

	// Append the record to the log
	off, err := log.Append(append)
	require.NoError(t, err)
	require.Equal(t, uint64(0), off)

	// Use the log's Reader to read back the data
	reader := log.Reader()
	b, err := io.ReadAll(reader)
	require.NoError(t, err)

	// Prefix length of each entry. Rule is set by index and store wordLength.
	var wordLength uint64 = 8
	read := &api.Record{}
	err = proto.Unmarshal(b[wordLength:], read)

	// Unmarshal the data read from the log back into a record
	require.NoError(t, err)
	require.Equal(t, append.Value, read.Value)

	// Verify the original and read records are equal
	require.Equal(t, append.Value, read.Value, "Read value should match the original appended value.")
}
