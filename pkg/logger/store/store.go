package store

import (
	"bufio"
	"encoding/binary"
	"errors"
	"os"
	"sync"
)

var (
	enc        = binary.BigEndian
	wordLength = 8
)

// These options are good to start with
// Will look into other options as time moves on.
// Options like:
//	- Asynchronous Writing
//	- Compression
//	- File Rollover
//	- Auto-Flush Interval

type Options struct {
	BufferSize uint64
	File       *os.File
	FilePath   string
	IsOpen     bool
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
		BufferSize: 4096,            // Default buffer size
		File:       nil,             // nil pointer
		FilePath:   "./default.txt", // destination of temp generate
	}
}

// Set the file for the store to write logs to
func WithFile(f *os.File) StoreOptions {
	return func(opts *Options) {
		opts.File = f
	}
}

// Specifies the file path for the store's backing file
func WithFilePath(path string) StoreOptions {
	return func(opts *Options) {
		opts.FilePath = path
	}
}

// Set the size of the buffer used by the store
func WithBufferSize(size uint64) StoreOptions {
	return func(opts *Options) {
		opts.BufferSize = size
	}
}

// Creates a new store with the given options.
// It initializes a store with a buffer of the specified size and associates it with the provided file, if any.
// The function applies a series of StoreOptions functions to configure the store.
func NewStore(optFns ...StoreOptions) (filestore *Store, err error) {
	// Initialize with default options.
	opts := DefaultOptions()

	// Apply each provided option to the default options
	for _, fn := range optFns {
		fn(opts)
	}

	var file *os.File

	// Check if a custom file is provided in options
	if opts.File == nil {
		// Open the default file, create if it does not exist, and set it to append mode
		file, err = os.OpenFile(opts.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err // Return an error if the file cannot be opened or created
		}
	} else if opts.File != nil {
		// If the file is already open, check if it's usable
		if _, err := opts.File.Stat(); err != nil {
			return nil, err
		}

		// Optionally, reset the file's offset or ensure it's ready for use
		if _, err := opts.File.Seek(0, os.SEEK_END); err != nil {
			return nil, err
		}

		file = opts.File
	}

	// Create a buffered writer with the specified buffer size
	buf := bufio.NewWriterSize(file, int(opts.BufferSize))

	// Return a new Store instance
	return &Store{
		File: file,
		buf:  buf,
		mu:   sync.Mutex{},
		size: 0, // Initial store size is 0.
	}, nil

}

func (store *Store) Append(entry []byte) (size uint64, pos uint64, err error) {
	// Lock the store to prevent concurrent writes
	store.mu.Lock()
	defer store.mu.Unlock()

	// Position holds the current size of the store,
	// which is also the position where new data will be appended.
	position := store.size

	// Write the length of the page first as a prefix
	// This length prefix allows for knowing how much to read during retrieval
	if err := binary.Write(store.buf, enc, uint64(len(entry))); err != nil {
		return 0, 0, err
	}

	// Write the contents of the page to the store
	written, err := store.buf.Write(entry)
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

func (store *Store) Read(pos uint64) ([]byte, error) {
	// Lock the store to prevent concurrent reads
	store.mu.Lock()
	defer store.mu.Unlock()

	// Even if the client gave the option to not have a file initially,
	// there still must be a file to read from they they have designated
	if store.File == nil {
		return nil, errors.New("store file is nil")
	}

	// Check if the file actually exists
	fileInfo, err := store.File.Stat()
	if err != nil {
		return nil, err
	}

	// Check if the position is within the file bounds
	if int64(pos) >= fileInfo.Size() || int64(pos)+int64(wordLength) > fileInfo.Size() {
		return nil, errors.New("position out of file bounds")
	}

	// Read the size of the data first
	sizeBuffer := make([]byte, wordLength)
	if _, err := store.File.ReadAt(sizeBuffer, int64(pos)); err != nil {
		return nil, err
	}

	// Decode the size using the same encoding used in writing
	dataSize := enc.Uint64(sizeBuffer)

	// Allocate a slice to hold the actual data
	data := make([]byte, dataSize)

	// Read the actual data
	if _, err := store.File.ReadAt(data, int64(pos)+int64(wordLength)); err != nil {
		return nil, err
	}

	return data, nil
}

func (store *Store) Close() error {
	// Lock the store to prevent any more actions
	store.mu.Lock()
	defer store.mu.Unlock()

	// First, flush any data in the buffer to ensure all
	// written data is saved to the file.
	if err := store.buf.Flush(); err != nil {
		return err
	}
	store.buf = nil

	// Close the file after flushing the buffer
	//This ensures that all buffered data is safely written to the file
	if err := store.File.Close(); err != nil {
		return err
	}

	return nil
}
