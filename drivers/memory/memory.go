package memory

import (
	"github.com/alash3al/redix/configparser"
	"github.com/alash3al/redix/driver"
	"github.com/armon/go-radix"
)

type Engine struct {
	memtable *radix.Tree
	config   *configparser.Config
}

func (e Engine) Open(config *configparser.Config) (driver.Driver, error) {
	return &Engine{
		memtable: radix.New(),
		config:   config,
	}, nil
}

func (e Engine) Close() error {
	return nil
}

func (e *Engine) Put(entry *driver.Entry) error {
	if _, ok := e.memtable.Insert(entry.Key, entry); !ok {
		return driver.ErrUnableToInsert
	}

	return nil
}

func (e *Engine) Delete(key string) error {
	_, deleted := e.memtable.Delete(key)

	if !deleted {
		return driver.ErrUnableToDelete
	}

	return nil
}

func (e *Engine) DeletePrefix(prefix string) error {
	e.memtable.DeletePrefix(prefix)

	return nil
}

func (e Engine) Get(key string) (*driver.Entry, error) {
	entryInterface, found := e.memtable.Get(key)
	if !found {
		return nil, driver.ErrKeyNotFound
	}

	entry := entryInterface.(*driver.Entry)

	return entry, nil
}

func (e Engine) Walk(fn func(*driver.Entry) bool) error {
	e.memtable.Walk(func(key string, val interface{}) bool {
		return fn(val.(*driver.Entry))
	})

	return nil
}

func (e Engine) WalkPrefix(prefix string, fn func(*driver.Entry) bool) error {
	e.memtable.WalkPrefix(prefix, func(key string, val interface{}) bool {
		return fn(val.(*driver.Entry))
	})

	return nil
}
