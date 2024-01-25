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
	// Create a new store with nil file
	store, err := NewStore(nil)
	if err != nil {
		t.Errorf("Expected no error when creating store with nil file, got: %v", err)
	}

	// Verify that the file in the store is nil initially
	if store.File != nil {
		t.Errorf("Expected nil file in store, got: %v", store.File)
	}

	// Create a temporary file
	tmpfile, err := os.CreateTemp("", "*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Clean up
	defer os.Remove(tmpfile.Name())

	// Assign file
	store.File = tmpfile

	// Verify that the file in the store is correctly assigned
	if store.File != tmpfile {
		t.Errorf("Expected file in store to be %v, got: %v", tmpfile, store.File)
	}
}
