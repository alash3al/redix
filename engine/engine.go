package engine

import (
	"sync"
	"time"

	"github.com/alash3al/redix/configparser"
	"github.com/alash3al/redix/driver"
	"github.com/alash3al/redix/driver/memory"
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
	memtable driver.Driver

	onChangeFunc []OnChangeFunc

	sync.RWMutex
}

// New initialize the engine
func New(config *configparser.Config) (*Engine, error) {
	mem, err := driver.Open(memory.Name, config)
	if err != nil {
		return nil, err
	}

	return &Engine{
		memtable:     mem,
		onChangeFunc: make([]OnChangeFunc, 0),
	}, nil
}

// Put insert/update the specified entry
// P.S: this doesn't handle locking, the caller should utilize the underlying sync.RWMutex when needed.
func (e *Engine) Put(entry *driver.Entry) (*driver.Entry, error) {
	currentTime := time.Now()

	oldEntry, err := e.memtable.Get(string(entry.Key))
	if err != nil && err != driver.ErrKeyNotFound {
		return nil, err
	}

	op := OpUpdate

	if oldEntry == nil {
		oldEntry = entry
		oldEntry.CreatedAt = &currentTime
		op = OpCreate
	}

	oldEntry.UpdatedAt = &currentTime
	oldEntry.Value = entry.Value

	if err := e.memtable.Put(oldEntry); err != nil {
		return nil, err
	}

	// TODO use a goroutine pool?
	go (func() {
		e.publishChange(entry.Key, op)
	})()

	return entry, nil
}

// Delete deletes the specified key
func (e *Engine) Delete(key string) error {
	if err := e.memtable.Delete(key); err != nil {
		return err
	}

	// TODO use a goroutine pool?
	go (func() {
		e.publishChange(key, OpDeleteKey)
	})()

	return nil
}

// DeletePrefix delete the subtree under a prefix
func (e *Engine) DeletePrefix(prefix string) error {
	if err := e.memtable.DeletePrefix(prefix); err != nil {
		return err
	}

	// TODO use a goroutine pool?
	go (func() {
		e.publishChange(prefix, OpDeletePrefix)
	})()

	return nil
}

// Get return the value of the specified key
func (e *Engine) Get(key string) (*driver.Entry, error) {
	entry, err := e.memtable.Get(key)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

// Walk is used to walk the tree
func (e *Engine) Walk(fn func(*driver.Entry) bool) error {
	return e.memtable.Walk(fn)
}

// WalkPrefix is used to walk the tree under a prefix
func (e *Engine) WalkPrefix(prefix string, fn func(*driver.Entry) bool) error {
	return e.memtable.WalkPrefix(prefix, fn)
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
