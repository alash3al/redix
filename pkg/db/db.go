package db

import (
	"fmt"
	"path/filepath"
	"sync"
	"time"

	"github.com/tidwall/gjson"

	"github.com/alash3al/redix/pkg/db/driver"
)

var (
	databases = sync.Map{}
)

const (
	defaultAsyncModeWindowDuration = time.Second * 5
)

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

// DB represents datastore
type DB struct {
	provider driver.IDriver

	name   string
	driver string
	writes []driver.Entry

	asyncModeEnabled bool
	asyncModeWindow  time.Duration

	l *sync.RWMutex
}

// Open a database
func Open(provider, dirname, dbname string, opts gjson.Result) (*DB, error) {
	asyncModeEnabled := opts.Get("async_mode.enable").Bool()
	asyncModeWindowStr := opts.Get("async_mode.window").String()

	var asyncModeWindowDur time.Duration
	var err error

	if asyncModeEnabled && asyncModeWindowStr != "" {
		asyncModeWindowDur, err = time.ParseDuration(asyncModeWindowStr)
		if err != nil {
			return nil, err
		}
	}

	if asyncModeWindowStr == "" {
		asyncModeWindowDur = defaultAsyncModeWindowDuration
	}

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
		provider:         driverInstance,
		name:             dbname,
		driver:           provider,
		writes:           []driver.Entry{},
		l:                &sync.RWMutex{},
		asyncModeEnabled: asyncModeEnabled,
		asyncModeWindow:  asyncModeWindowDur,
	}

	databases.Store(dbkey, db)

	go (func() {
		t := time.NewTicker(db.asyncModeWindow)
		defer t.Stop()

		for range t.C {
			db.l.Lock()
			if len(db.writes) < 1 {
				db.l.Unlock()
				continue
			}

			batch := make([]driver.Entry, len(db.writes))
			copy(batch, db.writes)
			db.writes = db.writes[:0]
			db.l.Unlock()

			db.provider.Batch(batch)
		}
	})()

	return db, nil
}

// Name return the current database name
func (db DB) Name() string {
	return db.name
}

// Put writes a key - value pair into the database
func (db *DB) Put(entry driver.Entry) error {
	if entry.WriteMerger != nil {
		oldVal, _ := db.provider.Get(entry.Key)
		entry.Value = entry.WriteMerger(oldVal, entry.Value)
	}

	if !db.asyncModeEnabled {
		if entry.Value == nil {
			return db.provider.Delete(entry.Key)
		}
		return db.provider.Put(entry)
	}

	db.l.Lock()
	defer db.l.Unlock()

	db.writes = append(db.writes, entry)

	return nil
}

// Get fetches a document using its primary key
func (db DB) Get(key []byte) ([]byte, error) {
	data, err := db.provider.Get(key)
	if err == driver.ErrKeyExpired {
		db.l.Lock()
		defer db.l.Unlock()

		db.writes = append(db.writes, driver.Entry{Key: key})

		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Has whether a key exists or not
func (db DB) Has(key []byte) (bool, error) {
	val, err := db.Get(key)
	return val != nil, err
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
