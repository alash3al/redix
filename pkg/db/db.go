package db

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/alash3al/redix/pkg/db/driver"
)

// DB represents datastore
type DB struct {
	provider driver.IDriver
	name     string
	driver   string
	buffer   []driver.Pair
	l        *sync.RWMutex
}

// Open a database
func Open(provider, dirname string, dbname string, opts map[string]interface{}) (*DB, error) {
	dbkey := fmt.Sprintf("%s_%s", provider, dbname)

	if dbInterface, loaded := databases.Load(dbkey); loaded {
		return dbInterface.(*DB), nil
	}

	driverImpl, err := driver.Get(provider)
	if err != nil {
		return nil, err
	}

	driverInstance, err := driverImpl.Open(filepath.Join(dirname, dbname), opts)
	if err != nil {
		return nil, err
	}

	db := &DB{
		provider: driverInstance,
		name:     dbname,
		driver:   provider,
		buffer:   []driver.Pair{},
		l:        &sync.RWMutex{},
	}

	databases.Store(dbkey, db)

	go (func() {
		for {
			time.Sleep(time.Second * 1)

			db.l.Lock()

			newBuffer := make([]driver.Pair, len(db.buffer))
			copy(newBuffer, db.buffer)
			db.buffer = nil

			db.l.Unlock()

			for _, pair := range newBuffer {
				db.commitSync(pair)
			}
		}
	})()

	return db, nil
}

// CloseAll close all opened dbs
func CloseAll() {
	databases.Range(func(_, v interface{}) bool {
		db, ok := v.(*DB)
		if ok {
			db.Close()
		}
		return true
	})
}

// Name return the current database name
func (db DB) Name() string {
	return db.name
}

// Put writes a key - value pair into the database
func (db *DB) Put(pair driver.Pair) error {
	if pair.Async {
		return db.commitAsync(pair)
	}

	return db.commitSync(pair)
}

// Get fetches a document using its primary key
func (db DB) Get(key []byte) (*driver.Pair, error) {
	data, err := db.provider.Get(key)
	if err != nil {
		return nil, err
	}

	pair, err := driver.DecodePair(data)
	if err != nil {
		return nil, err
	}

	if pair == nil {
		return nil, nil
	}

	if pair.TTL > 0 && time.Now().Sub(pair.CommitedAt).Seconds() >= float64(pair.TTL) {
		db.commitAsync(driver.Pair{Key: pair.Key})
		return nil, nil
	}

	return pair, nil
}

// Has whether a key exists or not
func (db DB) Has(key []byte) (bool, error) {
	return db.provider.Has(key)
}

// Scan scans the db
func (db DB) Scan(opts driver.ScanOpts) {
	if opts.Scanner == nil {
		return
	}

	db.provider.Scan(opts)
}

// Close the database
func (db DB) Close() error {
	return db.provider.Close()
}

// commitSync puts new document into the storage [sync]
func (db DB) commitSync(pair driver.Pair) error {
	if nil == pair.Value {
		return db.provider.Delete(pair.Key)
	}

	if pair.TTL > 0 {
		pair.CommitedAt = time.Now()
	}

	if pair.WriteMerger != nil {
		if oldPair, err := db.Get(pair.Key); err != nil {
			pair.Value = pair.WriteMerger(*oldPair, pair)
		} else {
			return err
		}
	}

	value, err := driver.EncodePair(pair)
	if err != nil {
		return err
	}

	return db.provider.Put(pair.Key, value)
}

// commitAsync puts new document into the storage [async]
func (db *DB) commitAsync(pair driver.Pair) error {
	db.l.Lock()
	db.buffer = append(db.buffer, pair)
	db.l.Unlock()

	return nil
}
