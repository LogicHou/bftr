package factory

import (
	"sync"
)

var (
	providersMu sync.RWMutex
)

func Register(name string) {
	println("todo")
}

func New(providerName string) {
	println("todo")
}
