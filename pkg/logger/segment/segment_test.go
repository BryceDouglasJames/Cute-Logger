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

	want := &api.Record{Value: []byte("test value")}

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

	defer os.RemoveAll(dir)
	defer os.RemoveAll(dir2)
	defer func() {
		require.NoError(t, seg.Close())
	}()

}
