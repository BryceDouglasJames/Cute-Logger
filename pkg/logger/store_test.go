package logger

import (
	"os"
	"testing"
)

func TestNewStoreWithValidFile(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Clean up
	defer os.Remove(tmpFile.Name())

	// Test creating a new store with the temp file
	_, err = NewStore(tmpFile, WithBufferSize(4096)) // 4 KB page
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
	tmpFile, err := os.CreateTemp("", "*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Clean up
	defer os.Remove(tmpFile.Name())

	// Assign file
	store.File = tmpFile

	// Verify that the file in the store is correctly assigned
	if store.File != tmpFile {
		t.Errorf("Expected file in store to be %v, got: %v", tmpFile, store.File)
	}
}
