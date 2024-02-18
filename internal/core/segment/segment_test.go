package segment

import (
	"io"
	"os"
	"testing"

	api "github.com/BryceDouglasJames/Cute-Logger/api"
	"github.com/stretchr/testify/require"
)

func TestNewSegment(t *testing.T) {
	dir, err := os.MkdirTemp("", "segment-test")
	require.NoError(t, err)

	want := &api.Record{
		Value:  []byte("test value"),
		Offset: 0,
	}

	entryLength := uint64(12)

	// Define options for the new segment
	opts := []SegmentOptions{
		WithFilePath(dir),
		WithMaxStoreBytes(1024),
		WithMaxIndexBytes(entryLength * 3),
		WithInitialOffset(0),
	}

	// Create a new segment with the specified options
	seg, err := NewSegment(opts...)
	require.NoError(t, err)

	// Verify the segment is initialized with expected values
	require.Equal(t, uint64(0), seg.nextOffset)

	// Test appending and reading records from the segment
	for i := uint64(0); i < 3; i++ {
		offset, err := seg.Append(want)
		require.NoError(t, err)
		require.Equal(t, seg.baseOffset+i, offset)

		got, err := seg.Read(offset)
		require.NoError(t, err)
		require.Equal(t, want.Value, got.Value)
	}

	// Test the segment reaches its max capacity
	_, err = seg.Append(want)
	require.Equal(t, io.EOF, err)

	// Adjust the configuration to test different capacities
	dir2, err := os.MkdirTemp("", "segment-test-2")
	require.NoError(t, err)

	opts = []SegmentOptions{
		WithFilePath(dir2),
		WithMaxStoreBytes(uint64(len(want.Value) * 3)),
		WithMaxIndexBytes(1024),
		WithInitialOffset(0),
	}

	// Recreate the segment with new options
	seg, err = NewSegment(opts...)
	require.NoError(t, err)

	defer func() {
		os.RemoveAll(dir)
	}()

	defer func() {
		os.RemoveAll(dir2)
	}()

	defer func() {
		require.NoError(t, seg.Close())
	}()

}

func TestSegmentIsFull(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "segment-isfull-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Setup segment with small capacities to quickly reach full status during testing
	seg, err := NewSegment(
		WithFilePath(tempDir),
		WithMaxStoreBytes(50),
		WithMaxIndexBytes(50),
		WithInitialOffset(0),
	)
	require.NoError(t, err)

	defer func() {
		require.NoError(t, seg.Close())
	}()

	testRecord := &api.Record{
		Value:  []byte("test"),
		Offset: 0,
	}

	// Append records until the segment is close to full
	for !seg.IsFull() {
		_, err := seg.Append(testRecord)
		if err != nil {
			t.Fatalf("Failed to append record to segment: %v", err)
		}
	}

	// Confirm the segment reports it is full
	require.True(t, seg.IsFull(), "Segment should report it is full")

	// Attempt to append another record and expect failure or specific behavior indicating the segment is full
	_, err = seg.Append(testRecord)
	if err == nil {
		t.Error("Expected error when appending to a full segment")
	}
}

func TestSegmentRemove(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "segment_remove_test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a new segment for testing
	segment, err := NewSegment(
		WithFilePath(tempDir),
		WithInitialOffset(0),
	)
	require.NoError(t, err)

	// Ensure the segment's files exist before attempting removal
	_, err = os.Stat(segment.index.File.Name())
	require.NoError(t, err, "Index file should exist before removal")
	_, err = os.Stat(segment.store.Name())
	require.NoError(t, err, "Store file should exist before removal")

	// Attempt to remove the segment
	err = segment.Remove()
	require.NoError(t, err, "Segment removal should not produce an error")

	// Verify that the segment's files have been removed
	_, err = os.Stat(segment.index.File.Name())
	require.Error(t, err, "Index file should not exist after removal")
	require.True(t, os.IsNotExist(err), "Error should indicate that the index file does not exist")

	_, err = os.Stat(segment.store.Name())
	require.Error(t, err, "Store file should not exist after removal")
	require.True(t, os.IsNotExist(err), "Error should indicate that the store file does not exist")
}
