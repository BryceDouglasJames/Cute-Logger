package logger

import (
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
	logFiles, err := os.ReadDir(l.Directory)
	if err != nil {
		return err
	}

	var startingOffsets []uint64
	for _, file := range logFiles {
		offsetString := strings.TrimSuffix(file.Name(), path.Ext(file.Name()))

		offset, _ := strconv.ParseUint(offsetString, 10, 0)
		startingOffsets = append(startingOffsets, offset)
	}

	sort.Slice(startingOffsets,
		func(i, j int) bool {
			return startingOffsets[i] < startingOffsets[j]
		},
	)

	for i := 0; i < len(startingOffsets); i++ {
		if err = l.newSegment(startingOffsets[i]); err != nil {
			return err
		}

		//startingOffset contains dup for index and store so we skip the dup
		i++
	}

	if l.segmentList == nil {
		if err = l.newSegment(0); err != nil {
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
