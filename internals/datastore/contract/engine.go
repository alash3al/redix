package contract

import (
	"errors"
	"time"
)

// Engine represents an Engine
type Engine interface {
	Open(string) error
	Close() error
	Write(*WriteInput) (*WriteOutput, error)
	Read(*ReadInput) (*ReadOutput, error)
	Iterate(*IteratorOpts) error
	Publish([]byte, []byte) error
	Subscribe([]byte, func([]byte) error) error
}

// WriteInput represents a PUT request
type WriteInput struct {
	Key             []byte
	Value           []byte
	Increment       bool
	Append          bool
	OnlyIfNotExists bool
	TTL             time.Duration
	KeepTTL         bool
}

// WriteOutput represents a PUT output
type WriteOutput struct {
	Value []byte
	TTL   time.Duration
}

// ReadInput represents a Get request
type ReadInput struct {
	Key    []byte
	Delete bool
}

// ReadOutput represents a Get output
type ReadOutput struct {
	Key    []byte
	Value  []byte
	Exists bool
	TTL    time.Duration
}

// IteratorOpts represents the itrator options
type IteratorOpts struct {
	Prefix   []byte
	Callback func(*ReadOutput) error
}

// global vars
var (
	ErrStopIterator = errors.New("STOP_ITERATOR")
)
