package driver

import (
	"sync"
	"time"
)

// providers a providers for available drivers
var (
	providersMap  = map[string]IDriver{}
	providersLock = &sync.RWMutex{}
)

// IDriver an interface describes a storage backend
type IDriver interface {
	Open(string, map[string]interface{}) (IDriver, error)
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

// Register register a new driver
func Register(name string, provider IDriver) error {
	providersLock.Lock()
	defer providersLock.Unlock()

	if providersMap[name] != nil {
		return ErrDriverAlreadyExists
	}

	providersMap[name] = provider

	return nil
}

// Get returns a driver from the registery
func Get(name string) (IDriver, error) {
	providersLock.Lock()
	defer providersLock.Unlock()

	if providersMap[name] == nil {
		return nil, ErrDriverNotFound
	}

	return providersMap[name], nil
}
