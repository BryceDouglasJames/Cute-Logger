package index

import (
	"errors"
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
	AutoCreate       bool
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
		AutoCreate:       true,
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

// Option to enable or disable automatic index file creation.
func WithAutoCreate(autoCreate bool) IndexOptions {
	return func(opts *Options) {
		opts.AutoCreate = autoCreate
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
		// We want to be careful if we decide to auto create the index files for data integrity and consistency sake.
		// So we will let that be an option.
		if opts.AutoCreate {
			// Attempt to open or create the file only if AutoCreate is true.
			newIndex.file, err = os.OpenFile(opts.FilePath, os.O_RDWR|os.O_CREATE, 0666)
			if err != nil {
				return nil, err
			}
		} else {
			// Attempt to open the file without creating it.
			newIndex.file, err = os.Open(opts.FilePath)
			if err != nil {
				return nil, err
			}
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

	// Attempt to memory-map the file if requested
	if opts.UseMemoryMapping {
		// Ensure the file descriptor supports the intended memory map protections.
		mmapProt := gommap.PROT_READ | gommap.PROT_WRITE
		mmapFlags := gommap.MAP_SHARED

		newIndex.mmap, err = gommap.Map(newIndex.file.Fd(), mmapProt, mmapFlags)
		if err != nil {
			return nil, err
		}
		newIndex.UseMemoryMapping = true
	}

	return newIndex, nil
}

func (i *Index) Close() error {
	// Check if mmap exists and is valid before attempting to sync
	if i.mmap != nil {
		if err := i.mmap.Sync(gommap.MS_SYNC); err != nil {
			return err
		}
	} else if i.UseMemoryMapping {
		return errors.New("something is very wrong index mmap should not be nil")
	}

	// Ensure file is synced and truncated properly
	if err := i.file.Sync(); err != nil {
		return err
	}
	if err := i.file.Truncate(int64(i.size)); err != nil {
		return err
	}

	// Close the file at the end
	if err := i.file.Close(); err != nil {
		return err
	}

	return nil
}
