package logger

import (
	"os"
	"testing"
)

func TestNewStoreWithValidFile(t *testing.T) {
	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Clean up
	defer os.Remove(tmpfile.Name())

	// Test creating a new store with the temp file
	_, err = NewStore(tmpfile, WithBufferSize(4096)) // 4 KB page
	if err != nil {
		t.Errorf("Failed to create new store: %v", err)
	}
}

func TestNewStoreWithNilFile(t *testing.T) {
	// Test creating a new store with nil file
	_, err := NewStore(nil)
	if err == nil {
		t.Errorf("Expected error when creating store with nil file, got none")
	}
}
