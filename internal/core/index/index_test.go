package index

import (
	"io"
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
	defer os.Remove(idx.File.Name())

	if idx.File == nil {
		t.Error("Expected default file to be set, got nil")
	}

	if idx.UseMemoryMapping {
		t.Error("Expected memory mapping to be disabled by default")
	}
}

func TestNewIndexWithMemoryMapping(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "0.index")
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

func TestIndexReadWrite(t *testing.T) {
	var i *Index

	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "0.index")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}

	// Attempt to create a new index with the memory-mapped file option
	i, err = NewIndex(WithFile(tmpFile), WithMemoryMapping(true))
	if err != nil {
		t.Fatalf("Failed to create index with memory mapping enabled: %v", err)
	}

	// Clean up
	defer os.Remove(tmpFile.Name())

	// Verify that memory mapping was enabled in the index
	if !i.UseMemoryMapping {
		t.Error("Expected memory mapping to be enabled.")
	}

	tests := []struct {
		name    string
		wantOff uint32
		wantPos uint64
		wantErr error
	}{
		{
			name:    "Write and Read First Entry",
			wantOff: 1,
			wantPos: 12,
			wantErr: nil,
		},
		{
			name:    "Write and Read Last Entry",
			wantOff: 2,
			wantPos: 24,
			wantErr: nil,
		},
		{
			name:    "Attempt to Read Beyond Written Data",
			wantOff: 0,
			wantPos: 0,
			wantErr: io.EOF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip writing for the "Attempt to Read Beyond Written Data" test case
			if tt.name != "Attempt to Read Beyond Written Data" {
				err := i.Write(tt.wantOff, tt.wantPos)
				if err != nil {
					t.Fatalf("Write() failed: %v", err)
				}
			}

			// Attempt to read back the entry/entries
			// 'off - 1' converts offset to index for Read
			out, pos, err := i.Read(int64(tt.wantOff - 1))
			if err != tt.wantErr {
				t.Errorf("%s: Read() error = %v, wantErr %v", tt.name, err, tt.wantErr)
				return
			}
			if tt.wantErr != nil {
				// If an error is expected, further checks are skipped
				return
			}
			if out != tt.wantOff {
				t.Errorf("%s: Read() gotOff = %v, want %v", tt.name, out, tt.wantOff)
			}
			if pos != tt.wantPos {
				t.Errorf("%s: Read() gotPos = %v, want %v", tt.name, pos, tt.wantPos)
			}
		})
	}
}

func TestIndexClose(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "0.index")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}

	// Clean up
	defer os.Remove(tmpFile.Name())

	// Initialize the Index with the temporary file
	i, err := NewIndex(WithFile(tmpFile), WithMemoryMapping(true), WithAutoCreate(false))
	if err != nil {
		t.Fatalf("Failed to create Index: %v", err)
	}

	// Close the Index and ensure no errors are returned
	if err := i.Close(); err != nil {
		t.Errorf("Failed to close Index: %v", err)
	}
}
