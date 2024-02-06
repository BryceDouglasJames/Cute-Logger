package segment

import (
	"github.com/BryceDouglasJames/Cute-Logger/pkg/logger/index"
	"github.com/BryceDouglasJames/Cute-Logger/pkg/logger/store"
)

type Segment struct {
	store      *store.Store
	index      *index.Index
	baseOffset uint64
	nextOffet  uint64
}

//TODO: Use gRPC for records so you can use them for segments
