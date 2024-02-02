package index

import (
	"os"

	"github.com/tysonmote/gommap"
)

var (
	offset      uint64 = 4
	wordLength  uint64 = 8
	entryLength        = offset + wordLength
)

type Options struct {
	File             *os.File
	FilePath         string
	UseMemoryMapping bool
}

// Represents a function that applies configuration options to an Options instance
type IndexOptions func(*Options)

type Index struct {
	file             *os.File
	size             uint64
	mmap             gommap.MMap
	UseMemoryMapping bool
}

// Default settings for store
func DefaultOptions() *Options {
	return &Options{
		File:             nil,             // nil pointer
		FilePath:         "./default.txt", // destination of temp generate
		UseMemoryMapping: false,
	}
}

// Set the file to be indexed
func WithFile(f *os.File) IndexOptions {
	return func(opts *Options) {
		opts.File = f
	}
}

// Specifies the file path for the store's backing file
func WithFilePath(path string) IndexOptions {
	return func(opts *Options) {
		opts.FilePath = path
	}
}

// Enables or disables memory mapping for the index file.
func WithMemoryMapping(use bool) IndexOptions {
	return func(opts *Options) {
		opts.UseMemoryMapping = use
	}
}

func NewIndex(optFns ...IndexOptions) (*Index, error) {
	// Initialize with default options.
	opts := DefaultOptions()

	// Apply each option to the Options struct
	for _, option := range optFns {
		option(opts)
	}

	var err error
	newIndex := &Index{}

	// Check if a custom file is provided in options
	if opts.File == nil {
		// Open the default file, create if it does not exist, and set it to append mode
		newIndex.file, err = os.OpenFile(opts.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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

		newIndex.file = opts.File
	} else {
		// No file or file path provided
		return nil, os.ErrInvalid
	}

	// Get file info to set the size
	fi, err := newIndex.file.Stat()
	if err != nil {
		return nil, err
	}
	newIndex.size = uint64(fi.Size())

	// TODO: Add Segment Management

	// Memory-map the file if requested
	if opts.UseMemoryMapping {
		newIndex.mmap, err = gommap.Map(newIndex.file.Fd(), gommap.PROT_READ|gommap.PROT_WRITE, gommap.MAP_SHARED)
		if err != nil {
			return nil, err
		}
		newIndex.UseMemoryMapping = true
	} else {
		newIndex.UseMemoryMapping = false
	}

	return newIndex, nil
}
