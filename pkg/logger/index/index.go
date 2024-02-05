package index

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/tysonmote/gommap"
)

var (
	offset      uint64 = 4
	wordLength  uint64 = 8
	entryLength        = offset + wordLength

	enc = binary.BigEndian
)

type Options struct {
	File             *os.File
	FilePath         string
	UseMemoryMapping bool
	AutoCreate       bool
	MaxIndexBytes    uint64
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
		MaxIndexBytes:    1024,
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

// Sets the maximum number of bytes for the index file itself.
func WithMaxIndexBytes(maxIndexBytes uint64) IndexOptions {
	return func(opts *Options) {
		opts.MaxIndexBytes = maxIndexBytes
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
		opts.FilePath = opts.File.Name()
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

	fmt.Println(newIndex.file)

	// Truncate new index into index file
	if err = os.Truncate(newIndex.file.Name(), int64(opts.MaxIndexBytes)); err != nil {
		return nil, err
	}

	// Attempt to memory-map the file if requested
	//  *********** BE CAREFUL! ***********
	//  Map creates a new mapping in the virtual address space of the calling process.
	// 	May have unexpected bahvior depending on architecture
	if opts.UseMemoryMapping {
		// Ensure the file descriptor supports the intended memory map protections.
		mmapProt := gommap.PROT_READ | gommap.PROT_WRITE
		mmapFlags := gommap.MAP_SHARED | gommap.MAP_ANONYMOUS

		newMap, err := gommap.Map(newIndex.file.Fd(), mmapProt, mmapFlags)
		if err != nil {
			return nil, err
		}
		newIndex.UseMemoryMapping = true
		newIndex.mmap = newMap
	}

	return newIndex, nil
}

func (i *Index) Write(off uint32, pos uint64) error {
	// Check if there's enough space left in the memory-mapped file to write a new entry
	if uint64(len(i.mmap)) < i.size+entryLength {
		return io.EOF
	}

	// Write the offset value to the memory-mapped file at the current size position
	enc.PutUint32(i.mmap[i.size:i.size+offset], off)

	// Write the position value immediately after offset in the memory-mapped file
	enc.PutUint64(i.mmap[i.size+offset:i.size+entryLength], pos)

	// Increase size counter for index
	i.size += uint64(entryLength)

	return nil
}

func (i *Index) Read(in int64) (out uint32, pos uint64, err error) {
	// If the index size is 0, return EOF to indicate no entries can be read
	if i.size == 0 {
		return 0, 0, io.EOF
	}

	// If in is -1, calculate the index of the last entry. Otherwise, use in as the index
	if in == -1 {
		out = uint32((i.size / entryLength) - 1)
	} else {
		out = uint32(in)
	}

	// Calculate the byte position of the entry within the memory-mapped file
	pos = uint64(out) * entryLength

	// If the calculated position is beyond the size of the index, return EOF
	if i.size < pos+entryLength {
		return 0, 0, io.EOF
	}

	// Read the entry value and position from the memory-mapped file
	out = enc.Uint32(i.mmap[pos : pos+offset])
	pos = enc.Uint64(i.mmap[pos+offset : pos+entryLength])

	return out, pos, nil
}

func (i *Index) Close() error {
	// Check if mmap exists and is valid before attempting to sync
	if i.mmap != nil {
		if err := i.mmap.Sync(gommap.MS_SYNC); err != nil {
			return err
		}
	} else if len(i.mmap) == 0 {
		i.mmap = nil
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
