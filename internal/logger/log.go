package logger

import (
	"errors"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"

	api "github.com/BryceDouglasJames/Cute-Logger/api"
	seg "github.com/BryceDouglasJames/Cute-Logger/internal/core/segment"
)

type Log struct {
	mutex     sync.RWMutex
	Directory string

	activeSegment *seg.Segment
	segmentList   []*seg.Segment
}

func NewLog(dir string) (log *Log, err error) {
	l := &Log{
		Directory: dir,
	}

	return l, l.setup()
}

func (l *Log) setup() error {
	// Attempt to read the directory for any existing log files
	logFiles, err := os.ReadDir(l.Directory)
	if err != nil {
		return err
	}

	// Parse the starting offsets from the filenames of log files
	var startingOffsets []uint64
	for _, file := range logFiles {
		offsetString := strings.TrimSuffix(file.Name(), path.Ext(file.Name()))
		offset, _ := strconv.ParseUint(offsetString, 10, 0)
		if err != nil {
			return errors.New("failed to parse offset")
		}
		startingOffsets = append(startingOffsets, offset)
	}

	// Sort the offsets to ensure segments are processed in order.
	sort.Slice(startingOffsets,
		func(i, j int) bool {
			return startingOffsets[i] < startingOffsets[j]
		},
	)

	// Create segments for each starting offset.
	// Skip every other offset since they are duplicated for index and store.
	for i := 0; i < len(startingOffsets); i += 2 {
		if err = l.newSegment(startingOffsets[i]); err != nil {
			return err
		}
	}

	// If no segments were found, initialize a new segment at offset 0
	if len(l.segmentList) == 0 {
		if err := l.newSegment(0); err != nil {
			return err
		}
	}

	return nil
}

func (l *Log) Append(record *api.Record) (offset uint64, err error) {
	// Protect Read/Write
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Append record to active segment
	off, err := l.activeSegment.Append(record)
	if err != nil {
		return 0, err
	}

	// If the active segment is now full, create a new one.
	if l.activeSegment.IsFull() {
		err = l.newSegment(off + 1)
	}

	return off, err
}

func (l *Log) Read(offset uint64) (*api.Record, error) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	// Declare a pointer to hold the segment containing the offset
	var s *seg.Segment

	// Iterate through all segments to find the one containing the offset
	for _, segment := range l.segmentList {

		// Check if the current segment's range includes the offset
		if segment.BaseOffset() <= offset && offset < segment.NextOffset() {
			s = segment
			break
		}
	}

	// Check if segment is found or the found segment's next offset is not greater than the given offset
	if s == nil || s.NextOffset() <= offset {
		return nil, errors.New("offset is out of range when reading segments")
	}

	return s.Read(offset) // Read and return the record from the found segment
}

func (l *Log) Truncate(lowest uint64) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Prepare a slice to hold segments that are not removed
	var retainedSegments []*seg.Segment

	// Iterate over all segments in the log
	for _, s := range l.segmentList {

		// Check if the segment's next offset is before the truncation threshold
		if s.NextOffset() <= lowest+1 {

			// If so, attempt to remove the segment from the filesystem
			if err := s.Remove(); err != nil {
				return err
			}

			// Skip appending this segment to the retained segments
			continue
		}
		// If the segment is beyond the truncation threshold, retain it
		retainedSegments = append(retainedSegments, s)
	}

	// Update the log's segments to only include those that have been retained
	l.segmentList = retainedSegments

	return nil
}

func (l *Log) Close() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	// Iterate through all segments and attempt to close them.
	for _, seg := range l.segmentList {
		if err := seg.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (l *Log) Delete() error {
	// Close all segments to ensure that all resources are released.
	if err := l.Close(); err != nil {
		return errors.New("failed to close segments")
	}

	// Remove the log directory along with all its contents.
	if err := os.RemoveAll(l.Directory); err != nil {
		return errors.New("failed to remove log directory")
	}

	return nil
}

func (l *Log) Reset() error {
	// Delete the current log data, including all files and directories
	if err := l.Delete(); err != nil {
		return err
	}

	// Ensure the log directory is recreated after deletion
	if err := os.MkdirAll(l.Directory, 0755); err != nil {
		return errors.New("failed to recreate log directory")
	}

	// Reinitialize the log to its initial state
	return l.setup()
}

func (l *Log) newSegment(offset uint64) error {
	s, err := seg.NewSegment(
		seg.WithFilePath(l.Directory),
		seg.WithInitialOffset(offset),
	)

	if err != nil {
		return err
	}

	l.segmentList = append(l.segmentList, s)
	l.activeSegment = s
	return nil
}
