package segment_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BryceDouglasJames/Cute-Logger/pkg/logger/segment"
)

func TestNewSegment(t *testing.T) {
	// Use the system's temporary directory for the test
	tempDir := os.TempDir()

	// Define options for the new segment
	opts := []segment.SegmentOptions{
		segment.WithFilePath(tempDir),
		segment.WithMaxStoreBytes(1024),
		segment.WithMaxIndexBytes(512),
		segment.WithInitialOffset(0),
	}

	// Attempt to create a new segment with the specified options
	seg, err := segment.NewSegment(opts...)
	if err != nil {
		// If there is an error in creating the segment, fail the test immediately
		t.Fatalf("Failed to create new segment: %v", err)
	}

	// Verify that the store file has been created successfully
	storePath := filepath.Join(tempDir, "0.store")
	if _, err := os.Stat(storePath); os.IsNotExist(err) {
		// If the store file does not exist, report an error.
		t.Errorf("Store file was not created")
	}

	// Verify that the index file has been created successfully
	indexPath := filepath.Join(tempDir, "0.index")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		// If the index file does not exist, report an error.
		t.Errorf("Index file was not created")
	}

	// Close the segment and ensure that no error occurs
	if err := seg.Close(); err != nil {
		t.Errorf("There was an issue closing the segment: %v", err)
	}
}
