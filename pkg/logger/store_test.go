package logger

import (
	"os"
	"reflect"
	"testing"
)

func TestNewStoreWithValidFileFirst(t *testing.T) {
	// Define expected buffer size
	expectedBufferSize := 4096

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Schedule cleanup of the file
	defer os.Remove(tmpFile.Name())

	// Ensure the file is closed after setup
	defer tmpFile.Close()

	// Test creating a new store with the temp file and a specified buffer size
	store, err := NewStore(WithFile(tmpFile), WithBufferSize(4096))
	if err != nil {
		t.Errorf("Failed to create new store: %v", err)
	}

	// Check if the buffer size is set as expected
	if store.buf.Size() != expectedBufferSize {
		t.Errorf("Expected buffer size to be %d, got %d", expectedBufferSize, store.buf.Size())
	}

	// Validate the file association
	if store.File != tmpFile {
		t.Errorf("Store is not associated with the correct file")
	}

	// Check the initial size of the store
	if store.size != 0 {
		t.Errorf("Expected initial store size to be 0, got %d", store.size)
	}
}

func TestNewStoreWithNilFileFirst(t *testing.T) {
	// Create a new store without specifying a file (should use default file settings)
	store, err := NewStore()
	if err != nil {
		t.Fatalf("Failed to create store with default file settings: %v", err)
	}

	// Verify that a default file is created and assigned in the store
	if store.File == nil {
		t.Fatalf("Expected a default file in store, got nil")
	}

	// We should further validate the default file (e.g., name, path, etc.)
	//...
	//...
	//...

	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Clean up
	defer os.Remove(tmpFile.Name())

	// Assign the temporary file to the store
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
	store, err := NewStore(WithFile(tmpfile))
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

func TestStoreRead(t *testing.T) {
	// Create a temporary file for testing.
	tmpfile, err := os.CreateTemp("", "store_read_test.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name()) // Clean up after the test.

	// Initialize a new Store with the temporary file.
	store, err := NewStore(WithFile(tmpfile))
	if err != nil {
		t.Fatalf("Failed to create new store: %v", err)
	}

	// Define a test page to append.
	testPage := []byte("test log data")

	// Append the test page to the store and capture the total written amount and position.
	totalWritten, pos, err := store.Append(testPage)
	if err != nil {
		t.Fatalf("Failed to append to store: %v", err)
	}

	// Check if the total written amount is correct.
	expectedTotalWritten := uint64(len(testPage)) + uint64(wordLength)
	if totalWritten != expectedTotalWritten {
		t.Errorf("Expected total written amount to be %d, got %d", expectedTotalWritten, totalWritten)
	}

	// Read the data back from the store.
	readData, err := store.Read(pos)
	if err != nil {
		t.Fatalf("Failed to read from store: %v", err)
	}

	// Verify that the read data matches the written data.
	if !reflect.DeepEqual(readData, testPage) {
		t.Errorf("Read data does not match written data. Got %v, want %v", readData, testPage)
	}
}

func TestStoreInitializationWithFilePath(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "example_*.log")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}

	// Clean up
	defer os.Remove(tmpFile.Name())

	// Close the file as NewStore will open it
	tmpFile.Close()

	store, err := NewStore(WithFilePath(tmpFile.Name()))
	if err != nil {
		t.Fatalf("Failed to create store with file path: %v", err)
	}

	if err := store.Close(); err != nil {
		t.Errorf("Failed to close store: %v", err)
	}
}
