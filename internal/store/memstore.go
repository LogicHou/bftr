package store

import (
	"sync"

	"github.com/LogicHou/bftr/store/factory"
)

func init() {
	factory.Register("mem", &MemStore{})
}

type MemStore struct {
	sync.RWMutex
}
