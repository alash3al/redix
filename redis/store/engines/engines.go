// Package engines contains main redis engines helpers
package engines

import (
	"fmt"
	"sync"

	"github.com/alash3al/redix/configparser"
	"github.com/alash3al/redix/redis/store"
)

var (
	engines      = map[configparser.DatabaseEngine]store.Store{}
	enginesMutex = &sync.RWMutex{}
)

// RegisterStorageEngine registers a new storage engine
func RegisterStorageEngine(name configparser.DatabaseEngine, engine store.Store) {
	enginesMutex.Lock()
	defer enginesMutex.Unlock()

	if _, exists := engines[name]; exists {
		panic(fmt.Errorf("engine %s already registered before", name))
	}

	engines[name] = engine
}

// OpenStorageEngine opens a storage engine
func OpenStorageEngine(config *configparser.Config) (store.Store, error) {
	enginesMutex.Lock()
	defer enginesMutex.Unlock()

	engine, exists := engines[config.Engine]
	if !exists {
		return nil, fmt.Errorf("engine %s is unknown", config.Engine)
	}

	return engine.Connect(config)
}
