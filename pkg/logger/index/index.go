package index

import (
	"os"

	"github.com/tysontate/gommap"
)

var (
	offset      uint64 = 4
	wordLength  uint64 = 8
	entryLength        = offset + wordLength
)

type index struct {
	file *os.File
	size uint64
	mmap gommap.MMap
}

func newIndex(f *os.File) (*index, error) {
	new_index := &index{
		file: f,
	}

	fileInfo, err := f.Stat()
	if err != nil {
		return nil, err
	}

	new_index.size = uint64(fileInfo.Size())

	new_index.mmap, err = gommap.Map(new_index.file.Fd(), gommap.PROT_READ|gommap.PROT_WRITE, gommap.MAP_SHARED)
	if err != nil {
		return nil, err
	}

	return new_index, nil
}
