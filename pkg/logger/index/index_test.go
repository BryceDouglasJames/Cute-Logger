package index

import (
	"os"
	"testing"
)

func TestNewIndexDefaultOptions(t *testing.T) {
	// No options specified
	idx, err := NewIndex()
	if err != nil {
		t.Fatalf("Failed to create index with default options: %v", err)
	}

	// Cleanup
	defer os.Remove(idx.file.Name())

	if idx.file == nil {
		t.Error("Expected default file to be set, got nil")
	}

	if idx.UseMemoryMapping {
		t.Error("Expected memory mapping to be disabled by default")
	}
}

func TestNewIndexWithMemoryMapping(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "example_*.log")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}

	// Close the file as NewStore will open it
	tmpFile.Close()

	// Create new index with memory mapped option
	idx, err := NewIndex(WithMemoryMapping(true), WithFilePath(tmpFile.Name()))
	if err != nil {
		t.Fatalf("Failed to create index with memory mapping enabled: %v", err)
	}

	// Clean up
	defer os.Remove(tmpFile.Name())

	if !idx.UseMemoryMapping {
		t.Error("Expected memory mapping to be enabled")
	}
}

func TestIndexClose(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "index_test")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}

	// Clean up
	defer os.Remove(tmpFile.Name())

	// Initialize the Index with the temporary file
	i, err := NewIndex(WithFile(tmpFile), WithMemoryMapping(true))
	if err != nil {
		t.Fatalf("Failed to create Index: %v", err)
	}

	// Close the Index and ensure no errors are returned
	if err := i.Close(); err != nil {
		t.Errorf("Failed to close Index: %v", err)
	}
}
