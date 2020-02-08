package driver

import (
	"sync"

	"github.com/tidwall/gjson"
)

// providers a providers for available drivers
var (
	providersMap  = map[string]IDriver{}
	providersLock = &sync.RWMutex{}
)

// IDriver an interface describes a storage backend
type IDriver interface {
	Open(name string, options gjson.Result) (IDriver, error)
	Put(Entry) error
	Get([]byte) ([]byte, error)
	Has([]byte) (bool, error)
	Delete([]byte) error
	Batch([]Entry) error
	Scan(ScanOpts)
	Close() error
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
