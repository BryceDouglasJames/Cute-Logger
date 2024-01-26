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

func TestStoreAppend(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "store_append_test.*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Clean up the file after the test
	defer os.Remove(tmpfile.Name())

	// Create a new store with the temporary file
	store, err := NewStore(tmpfile)
	if err != nil {
		t.Fatalf("Failed to create new store: %v", err)
	}

	// Define a test page to append
	testPage := []byte("test log data")

	// Append the test page to the store
	written, pos, err := store.Append(testPage)
	if err != nil {
		t.Fatalf("Failed to append to store: %v", err)
	}

	// Check if the returned position is correct (should be 0 for the first write)
	if pos != 0 {
		t.Errorf("Expected position 0, got %d", pos)
	}

	// Check if the number of written bytes is correct
	if written != uint64(len(testPage)+wordLength) {
		t.Errorf("Expected %d bytes written, got %d", len(testPage)+wordLength, written)
	}
}
