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
