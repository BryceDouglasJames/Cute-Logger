package segment_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/BryceDouglasJames/Cute-Logger/pkg/logger/segment"
)

func TestNewSegment(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "segment_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	opts := []segment.SegmentOptions{
		segment.WithFilePath(tempDir),
		segment.WithMaxStoreBytes(1024),
		segment.WithMaxIndexBytes(512),
		segment.WithInitialOffset(0),
	}

	seg, err := segment.NewSegment(opts...)
	if err != nil {
		t.Fatalf("Failed to create new segment: %v", err)
	}

	fmt.Println(seg)
	// Verify store and index files are created
	storePath := filepath.Join(tempDir, "0.store")
	if _, err := os.Stat(storePath); os.IsNotExist(err) {
		t.Errorf("Store file was not created")
	}

	indexPath := filepath.Join(tempDir, "0.index")
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Errorf("Index file was not created")
	}

}
