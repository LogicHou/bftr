package store

import "errors"

var (
	ErrNotFound = errors.New("not found")
	ErrExist    = errors.New("exist")
)

type Wsk struct {
	h  float64
	l  float64
	c  float64
	v  float64
	E  int64
	cm float64
}

type Store interface {
}
