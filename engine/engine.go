package engine

import (
	"fmt"
	"sync"
	"time"

	"github.com/alash3al/redix/driver"
	"github.com/armon/go-radix"
)

// ChangeOp represents the change currently happens to a key.
type ChangeOp string

const (
	OpCreate       ChangeOp = "create"
	OpUpdate       ChangeOp = "update"
	OpDeleteKey    ChangeOp = "deleteKey"
	OpDeletePrefix ChangeOp = "deletePrefix"
)

// OnChangeFunc executed when a key is changed.
type OnChangeFunc func(key string, op ChangeOp)

// Engine represents the core storage engine of the software.
type Engine struct {
	memtable *radix.Tree

	onChangeFunc []OnChangeFunc

	// TODO implement the persistent storage driver
	// driver   driver.Driver

	sync.RWMutex
}

// New initialize the engine
func New() (*Engine, error) {
	return &Engine{
		memtable:     radix.New(),
		onChangeFunc: make([]OnChangeFunc, 0),
	}, nil
}

// Put insert/update the specified entry
// P.S: this isn't thread-safe, the caller should utilize the underlying sync.RWMutex when needed.
func (e *Engine) Put(entry *driver.RawValueEntry) (*driver.RawValueEntry, error) {
	currentTime := time.Now()

	entry.UpdatedAt = &currentTime

	oldEntryInterface, found := e.memtable.Get(string(entry.Key))
	if !found {
		entry.CreatedAt = &currentTime
	} else {
		oldEntry := oldEntryInterface.(*driver.RawValueEntry)

		entry.CreatedAt = oldEntry.CreatedAt
	}

	_, inserted := e.memtable.Insert(string(entry.Key), entry)
	if !inserted {
		return nil, fmt.Errorf("unable to insert the specifed entry")
	}

	op := OpCreate

	if found {
		op = OpUpdate
	}

	// TODO use a goroutine pool?
	go (func() {
		e.publishChange(entry.Key, op)
	})()

	return entry, nil
}

// Delete deletes the specified key
func (e *Engine) Delete(key string) bool {
	_, deleted := e.memtable.Delete(key)

	if deleted {
		// TODO use a goroutine pool?
		go (func() {
			e.publishChange(key, OpDeleteKey)
		})()
	}

	return deleted
}

// DeletePrefix delete the subtree under a prefix
func (e *Engine) DeletePrefix(prefix string) bool {
	deletesCount := e.memtable.DeletePrefix(prefix)
	deleted := deletesCount > 0

	if deleted {
		// TODO use a goroutine pool?
		go (func() {
			e.publishChange(prefix, OpDeletePrefix)
		})()
	}

	return deleted
}

// Get return the value of the specified key
func (e *Engine) Get(key string) (*driver.RawValueEntry, bool) {
	entryInterface, found := e.memtable.Get(key)
	if !found {
		return nil, false
	}

	entry := entryInterface.(*driver.RawValueEntry)

	return entry, true
}

// Len returns the number of elements in the store
func (e *Engine) Len() int {
	return e.memtable.Len()
}

// Walk is used to walk the tree
func (e *Engine) Walk(fn func(*driver.RawValueEntry) bool) error {
	e.memtable.Walk(func(key string, val interface{}) bool {
		return fn(val.(*driver.RawValueEntry))
	})

	return nil
}

// WalkPrefix is used to walk the tree under a prefix
func (e *Engine) WalkPrefix(prefix string, fn func(*driver.RawValueEntry) bool) error {
	e.memtable.WalkPrefix(prefix, func(key string, val interface{}) bool {
		return fn(val.(*driver.RawValueEntry))
	})

	return nil
}

// Subscribe registers a watcher that will be notified when the tree is changed it returns the index of the new watcher
func (e *Engine) Subscribe(fn OnChangeFunc) int {
	e.Lock()
	defer e.Unlock()

	size := len(e.onChangeFunc)
	e.onChangeFunc = append(e.onChangeFunc, fn)

	return size + 1
}

// Unsubscribe unsubscribe the watcher of the specified index
func (e *Engine) Unsubscribe(idx int) {
	e.Lock()
	defer e.Unlock()

	if idx > (len(e.onChangeFunc) - 1) {
		return
	}

	e.onChangeFunc = append(e.onChangeFunc[0:idx], e.onChangeFunc[idx+1:]...)
}

func (e *Engine) publishChange(key string, op ChangeOp) {
	for _, fn := range e.onChangeFunc {
		fn(key, op)
	}
}
