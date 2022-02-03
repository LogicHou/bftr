package store

import (
	"sync"
)

func init() {
	println("todo")
}

type MemStore struct {
	sync.RWMutex
}
