package logger

import (
	"bufio"
	"encoding/binary"
	"os"
	"sync"
)

var (
	enc        = binary.BigEndian
	wordLength = 8
)

type Options struct {
	BufferSize uint64
	File       *os.File
}

// Represents a function that applies configuration options to an Options instance
type StoreOptions func(*Options)

type Store struct {
	mu   sync.Mutex
	buf  *bufio.Writer
	size uint64

	*os.File // File pointer to write logs to; if nil, the store will not be associated with a file initially
}

// Default settings for store
func DefaultOptions() *Options {
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
func NewStore(f *os.File, optFns ...StoreOptions) (filestore *Store, err error) {
	// Set options
	opts := DefaultOptions()
	for _, fn := range optFns {
		fn(opts)
	}

	// Create new store instance
	newStore := &Store{}

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
		newStore.size = fileSize

		buf := bufio.NewWriterSize(opts.File, int(fileSize))
		newStore.buf = buf

		// Return store with passed reference pointer
		return newStore, nil
	}

	// Use the buffer size specified in options for the buffered writer.
	// This allows for flexible configuration of the buffer size,
	// which can be optimized based on the expected file I/O workload.
	// For instance, setting it to the size of a standard page can optimize for page-aligned I/O.
	bufSize := opts.BufferSize
	buf := bufio.NewWriterSize(f, int(bufSize))
	newStore.buf = buf
	newStore.mu = sync.Mutex{}

	return newStore, nil

}

func (store *Store) Append(page []byte) (size uint64, pos uint64, err error) {
	// Lock the store to prevent concurrent writes
	store.mu.Lock()
	defer store.mu.Unlock()

	// Position holds the current size of the store,
	// which is also the position where new data will be appended.
	position := store.size

	// Write the length of the page first as a prefix
	// This length prefix allows for knowing how much to read during retrieval
	if err := binary.Write(store.buf, enc, uint64(len(page))); err != nil {
		return 0, 0, nil
	}

	// Write the contents of the page to the store
	written, err := store.buf.Write(page)
	if err != nil {
		return 0, 0, err
	}

	// Calculate the total number of bytes written (data + length prefix)
	totalWritten := uint64(written + wordLength)
	store.size += totalWritten

	// Flush the buffer to ensure all data is written to the underlying writer
	// Flushing is important to maintain data integrity
	if err := store.buf.Flush(); err != nil {
		return 0, 0, err
	}

	return totalWritten, position, nil
}
