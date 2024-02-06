package segment

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/BryceDouglasJames/Cute-Logger/pkg/logger/index"
	"github.com/BryceDouglasJames/Cute-Logger/pkg/logger/store"
)

type Segment struct {
	store      *store.Store
	index      *index.Index
	baseOffset uint64
	nextOffset uint64

	config *Options
}

type Options struct {
	FilePath      string
	MaxStoreBytes uint64
	MaxIndexBytes uint64
	InitialOffset uint64
}

// Default settings for segment
func DefaultOptions() *Options {
	return &Options{
		FilePath:      "./default.txt", // destination of temp generate
		MaxIndexBytes: 50 * 1024 * 1024,
		MaxStoreBytes: 10 * 1024 * 1024, // 10 MB
	}
}

// Represents a function that applies configuration options to an Options instance
type SegmentOptions func(*Options)

// WithFilePath sets the file path in the Options.
func WithFilePath(path string) SegmentOptions {
	return func(opts *Options) {
		opts.FilePath = path
	}
}

// WithMaxStoreBytes sets the maximum store bytes in the Options.
func WithMaxStoreBytes(maxBytes uint64) SegmentOptions {
	return func(opts *Options) {
		opts.MaxStoreBytes = maxBytes
	}
}

// WithMaxIndexBytes sets the maximum index bytes in the Options.
func WithMaxIndexBytes(maxBytes uint64) SegmentOptions {
	return func(opts *Options) {
		opts.MaxIndexBytes = maxBytes
	}
}

// WithInitialOffset sets the initial offset in the Options.
func WithInitialOffset(offset uint64) SegmentOptions {
	return func(opts *Options) {
		opts.InitialOffset = offset
	}
}

func NewSegment(optFns ...SegmentOptions) (*Segment, error) {
	// Initialize with default options.
	opts := DefaultOptions()

	// Apply each option to the Options struct
	for _, option := range optFns {
		option(opts)
	}

	// Validate mandatory file path
	if opts.FilePath == "" {
		return nil, errors.New("file path for segment is mandatory")
	}

	newSegment := &Segment{
		baseOffset: opts.InitialOffset,
		config:     opts,
	}

	// Construct the file path for the store and create/open the file
	storePath := path.Join(opts.FilePath, fmt.Sprintf("%d%s", opts.InitialOffset, ".store"))
	storeFile, err := os.OpenFile(storePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	// Initialize the store with the opened file
	if newSegment.store, err = store.NewStore(
		store.WithFile(storeFile),
	); err != nil {
		return nil, err
	}

	// Construct the file path for the index and create/open the file
	indexPath := path.Join(opts.FilePath, fmt.Sprintf("%d%s", opts.InitialOffset, ".index"))
	indexFile, err := os.OpenFile(
		indexPath,
		os.O_RDWR|os.O_CREATE,
		0644,
	)

	// Initialize the index with the opened file and configuration options
	if newSegment.index, err = index.NewIndex(
		index.WithFile(indexFile),
		index.WithMaxIndexBytes(opts.MaxIndexBytes),
	); err != nil {
		return nil, err
	}

	// Determine the next offset based on the last entry in the index, if any
	if off, _, err := newSegment.index.Read(-1); err != nil {
		newSegment.nextOffset = newSegment.baseOffset
	} else {
		newSegment.nextOffset = newSegment.baseOffset + uint64(off) + 1
	}

	return newSegment, nil

}
