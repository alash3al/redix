package driver

import (
	"time"
)

// Registry a registry for available drivers
var Registry = map[string]Interface{}

// Interface an interface describes a storage backend
type Interface interface {
	Open(string, map[string]interface{}) (Interface, error)
	Put([]byte, []byte) error
	Get([]byte) ([]byte, error)
	Has([]byte) (bool, error)
	// Batch([]Pair) error
	Delete([]byte) error
	Scan(ScanOpts)
	Close() error
}

// Pair represents a key - value pair
type Pair struct {
	Key         []byte
	Value       []byte
	TTL         int
	Async       bool
	WriteMerger func(Pair, Pair) []byte `msgpack:"-"`
	CommitedAt  time.Time
}
