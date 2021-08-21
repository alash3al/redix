package driver

import (
	"errors"
	"fmt"
	"sync"

	"github.com/alash3al/redix/internals/configparser"
)

var (
	ErrKeyNotFound    = errors.New("key not found")
	ErrUnableToInsert = errors.New("unable to insert")
	ErrUnableToDelete = errors.New("unable to delete")
)

type Driver interface {
	Open(*configparser.Config) (Driver, error)
	Close() error
	Put(*Entry) error
	Get(key string) (*Entry, error)
	Delete(key string) error
	DeletePrefix(prefix string) error
	Scan(opts ScanOpts, scanner func(*Entry) bool) error
}

type ScanOpts struct {
	Prefix           string
	StartingAfterKey string
	ResultLimit      int
}

var (
	drivers     = map[string]Driver{}
	driversLock = &sync.RWMutex{}
)

func Register(name string, driver Driver) error {
	driversLock.Lock()
	defer driversLock.Unlock()

	if _, found := drivers[name]; found {
		return fmt.Errorf("duplicate driver %s", name)
	}

	drivers[name] = driver

	return nil
}

func Open(name string, conf *configparser.Config) (Driver, error) {
	driversLock.RLock()
	defer driversLock.RUnlock()

	driver, found := drivers[name]
	if !found {
		return nil, fmt.Errorf("driver %s not found", name)
	}

	return driver.Open(conf)
}
