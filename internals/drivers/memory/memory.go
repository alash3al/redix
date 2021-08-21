package memory

import (
	"github.com/alash3al/redix/internals/configparser"
	"github.com/alash3al/redix/internals/driver"
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

func (e Engine) Scan(opts driver.ScanOpts, scanner func(*driver.Entry) bool) error {
	fetchedCount := 0

	scannerWrapper := func(key string, val interface{}) bool {
		if opts.ResultLimit > 0 && fetchedCount >= opts.ResultLimit {
			return true
		}

		if opts.StartingAfterKey != "" && opts.StartingAfterKey == key {
			return false
		}

		entry := val.(*driver.Entry)

		fetchedCount++

		return scanner(entry)
	}

	if opts.Prefix != "" {
		e.memtable.WalkPrefix(opts.Prefix, scannerWrapper)
	} else {
		e.memtable.Walk(scannerWrapper)
	}

	return nil
}
