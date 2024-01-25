package logger

import (
	"bufio"
	"os"
	"sync"
)

type Options struct {
	BufferSize uint64
	File       *os.File
}

// Represents a function that applies configuration options to an Options instance
type StoreOptions func(*Options)

type store struct {
	mu  sync.Mutex
	buf bufio.Writer

	*os.File // File pointer to write logs to; if nil, the store will not be associated with a file initially
}

// Default settings for store
func defaultOptions() *Options {
	return &Options{
		BufferSize: 4096, // Default buffer size
		File:       nil,  // nil pointer
	}
}

// Set the file for the store to write logs to
func WithFile(f *os.File) StoreOptions {
	return func(opts *Options) {
		opts.File = f
	}
}

// Set the size of the buffer used by the store.
func WithBufferSize(size uint64) StoreOptions {
	return func(opts *Options) {
		opts.BufferSize = size
	}
}

// Creates a new store with the given options.
// It initializes a store with a buffer of the specified size and associates it with the provided file, if any.
// The function applies a series of StoreOptions functions to configure the store.
func NewStore(f *os.File, optFns ...StoreOptions) (filestore *store, err error) {
	// Set options
	opts := defaultOptions()
	for _, fn := range optFns {
		fn(opts)
	}

	// Create new store instance
	newStore := &store{}

	if opts.File != nil {
		// Verify file from options
		file, err := os.Stat(opts.File.Name())
		if err != nil {
			return nil, err
		}
		newStore.File = opts.File

		// Add mutex to store
		newStore.mu = sync.Mutex{}

		// Buffer to fit file size
		fileSize := uint64(file.Size())
		buf := bufio.NewWriterSize(opts.File, int(fileSize))
		newStore.buf = *buf

		// Return store with passed reference pointer
		return newStore, nil
	}

	// Use the buffer size specified in options for the buffered writer.
	// This allows for flexible configuration of the buffer size,
	// which can be optimized based on the expected file I/O workload.
	// For instance, setting it to the size of a standard page can optimize for page-aligned I/O.
	bufSize := opts.BufferSize
	buf := bufio.NewWriterSize(f, int(bufSize))
	newStore.buf = *buf
	newStore.mu = sync.Mutex{}

	return newStore, nil

}

func (s *store) Append(page []byte) (size uint64, pos uint64, err error) { return 0, 0, nil }
