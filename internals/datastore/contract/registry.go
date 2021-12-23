package contract

import (
	"fmt"
	"sync"
)

var (
	engines     = map[string]Engine{}
	enginesLock = new(sync.RWMutex)
)

// Register adds the specified engine to the registry
func Register(name string, engine Engine) {
	enginesLock.Lock()
	defer enginesLock.Unlock()

	if _, exists := engines[name]; exists {
		panic(fmt.Errorf("duplicate driver name (%s)", name))
	}

	engines[name] = engine
}

// Open initialize an instance of the specified engine name + dsn
func Open(name string, dsn string) (Engine, error) {
	enginesLock.RLock()
	defer enginesLock.RUnlock()

	engine, exists := engines[name]
	if !exists {
		return nil, fmt.Errorf("unknown driver name (%s) specified", name)
	}

	return engine, engine.Open(dsn)
}

// Exists whether the specified engine name exists or not
func Exists(name string) bool {
	enginesLock.RLock()

	_, exists := engines[name]

	enginesLock.RUnlock()

	return exists
}
