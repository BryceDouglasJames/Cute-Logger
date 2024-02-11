package store

import (
	"os"
	"reflect"
	"testing"
)

func TestNewStoreWithValidFileFirst(t *testing.T) {
	// Define expected buffer size
	expectedBufferSize := 4096

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "*.store")
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
	if !reflect.DeepEqual(store.Buf.Size(), expectedBufferSize) {
		t.Errorf("Expected buffer size to be %d, got %d", expectedBufferSize, store.Buf.Size())
	}

	// Validate the file association
	if !reflect.DeepEqual(store.File, tmpFile) {
		t.Errorf("Store is not associated with the correct file")
	}

	// Check the initial size of the store
	if store.Size != 0 {
		t.Errorf("Expected initial store size to be 0, got %d", store.Size)
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
	tmpFile, err := os.CreateTemp("", "0.store")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Clean up
	defer os.Remove(tmpFile.Name())

	// Assign the temporary file to the store
	store.File = tmpFile

	// Verify that the file in the store is correctly assigned
	if !reflect.DeepEqual(store.File, tmpFile) {
		t.Errorf("Expected file in store to be %v, got: %v", tmpFile, store.File)
	}
}

func TestStoreAppend(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "store_append_test.*.store")
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
	if !reflect.DeepEqual(written, uint64(len(testPage)+wordLength)) {
		t.Errorf("Expected %d bytes written, got %d", len(testPage)+wordLength, written)
	}

	defer os.Remove("default.store")
}

func TestStoreRead(t *testing.T) {
	// Create a temporary file for testing
	tmpfile, err := os.CreateTemp("", "0.store")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Clean up after the test
	defer os.Remove(tmpfile.Name())

	// Initialize a new Store with the temporary file
	store, err := NewStore(WithFile(tmpfile))
	if err != nil {
		t.Fatalf("Failed to create new store: %v", err)
	}

	// Define a test page to append
	testPage := []byte("test log data")

	// Append the test page to the store and capture the total written amount and position
	totalWritten, pos, err := store.Append(testPage)
	if err != nil {
		t.Fatalf("Failed to append to store: %v", err)
	}

	// Check if the total written amount is correct
	expectedTotalWritten := uint64(len(testPage)) + uint64(wordLength)
	if totalWritten != expectedTotalWritten {
		t.Errorf("Expected total written amount to be %d, got %d", expectedTotalWritten, totalWritten)
	}

	// Read the data back from the store
	readData, err := store.Read(pos)
	if err != nil {
		t.Fatalf("Failed to read from store: %v", err)
	}

	// Verify that the read data matches the written data
	if !reflect.DeepEqual(readData, testPage) {
		t.Errorf("Read data does not match written data. Got %v, want %v", readData, testPage)
	}
}

func TestStoreClose(t *testing.T) {
	// Create a temporary file path
	tmpFile, err := os.CreateTemp("", "0.store")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	tmpFilePath := tmpFile.Name()

	// Clean up after the test
	defer os.Remove(tmpFilePath)

	// Initialize the Store with the file path
	store, err := NewStore(WithFile(tmpFile), WithBufferSize(4096))
	if err != nil {
		t.Fatalf("Failed to initialize store: %v", err)
	}

	// Append the test page to the store and capture the total written amount and position
	testEntry := []byte("test data")
	_, _, err = store.Append(testEntry)
	if err != nil {
		t.Fatalf("Failed to append to store: %v", err)
	}

	// Grab file state before close
	fileBefore, err := os.OpenFile(tmpFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to open up file %s: %v", tmpFilePath, err) // Return an error if the file cannot be opened or created
	}
	beforeInfo, err := fileBefore.Stat()
	if err != nil {
		t.Fatalf("Trouble grabbing file info from %s: %v", tmpFilePath, err)
	}
	beforeSize := beforeInfo.Size()

	// Ensure the store is closed properly to flush any buffered data
	if err := store.Close(); err != nil {
		t.Fatalf("Failed to close store: %v", err)
	}

	// Grab file state after close
	fileAfter, err := os.OpenFile(tmpFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to open up file %s: %v", tmpFilePath, err) // Return an error if the file cannot be opened or created
	}
	afterInfo, err := fileAfter.Stat()
	if err != nil {
		t.Fatalf("Trouble grabbing file info from %s: %v", tmpFilePath, err)
	}
	afterSize := afterInfo.Size()

	// Ensure store has flushed everything into the file
	if afterSize < beforeSize {
		t.Fatalf("The size of the file after store closure is not correct")
	}

}

func TestStoreInitializationWithFilePath(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "0.store")
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
