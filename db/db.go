package db

import (
	"fmt"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/alash3al/redix/db/driver"
)

var (
	countersPrefix  = []byte("internal::counters::")
	expirablePrefix = []byte("internal::expirable::")
)

// DB represents datastore
type DB struct {
	db           driver.Interface
	name         string
	driver       string
	writes       chan driver.KeyValue
	buffer       []driver.KeyValue
	useAsyncMode bool
	l            *sync.RWMutex
}

// Open a database
func Open(drivername, dirname string, dbname string, opts map[string]interface{}) (*DB, error) {
	dbkey := fmt.Sprintf("%s_%s", drivername, dbname)

	if dbInterface, loaded := databases.Load(dbkey); loaded {
		return dbInterface.(*DB), nil
	}

	driverImpl, exist := driver.Registry[drivername]
	if !exist {
		return nil, ErrDriverNotFound
	}

	driverInstance, err := driverImpl.Open(filepath.Join(dirname, dbname), opts)
	if err != nil {
		return nil, err
	}

	db := &DB{
		db:           driverInstance,
		name:         dbname,
		driver:       drivername,
		writes:       make(chan driver.KeyValue, 10000),
		buffer:       []driver.KeyValue{},
		l:            &sync.RWMutex{},
		useAsyncMode: opts["async"].(bool),
	}

	databases.Store(dbkey, db)

	go (func() {
		for {
			time.Sleep(time.Second * 1)

			db.l.Lock()
			newBuffer := make([]driver.KeyValue, len(db.buffer))
			copy(newBuffer, db.buffer)
			db.buffer = db.buffer[:0]

			db.l.Unlock()

			for _, pair := range newBuffer {
				db.putSync(pair.Key, pair.Value, pair.TTL)
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
func (db DB) Put(key, value []byte, ttl int) error {
	if db.useAsyncMode {
		return db.putAsync(key, value, ttl)
	}

	return db.putSync(key, value, ttl)
}

// Get fetches a document using its primary key
func (db DB) Get(key []byte) ([]byte, error) {
	data, err := db.db.Get(key)

	if err != nil {
		return nil, err
	}

	expirable, _ := db.db.Get(append(expirablePrefix, key...))
	if expirable == nil || len(expirable) < 1 {
		return data, nil
	}

	expires, _ := strconv.ParseInt(string(expirable), 10, 64)
	if expires > 0 && (time.Now().UnixNano() >= expires) {
		db.putAsync(key, nil, 0)
		return nil, nil
	}

	return data, nil
}

// Has whether a key exists or not
func (db DB) Has(key []byte) (bool, error) {
	return db.db.Has(key)
}

// Delete remove a key from the database [sync]
func (db DB) Delete(key []byte) error {
	return db.db.Delete(key)
}

// Scan scans the db
// TODO: handle custom value prefix like ttl value based
func (db DB) Scan(opts driver.ScanOpts) {
	if opts.Scanner == nil {
		return
	}

	db.db.Scan(opts)
}

// Close the database
func (db DB) Close() error {
	return db.db.Close()
}

// Put puts new document into the storage [sync]
func (db DB) putSync(key []byte, value []byte, ttl int) error {
	if nil == value {
		return db.db.Delete(key)
	}

	if ttl > 0 {
		expires := time.Now().Add(time.Duration(ttl) * time.Millisecond).UnixNano()
		db.db.Put(append(expirablePrefix, key...), []byte(strconv.FormatInt(expires, 10)))
	}

	return db.db.Put(key, value)
}

// PutAsync puts new document into the storage [async]
func (db *DB) putAsync(key []byte, value []byte, ttl int) error {
	db.l.Lock()
	db.buffer = append(db.buffer, driver.KeyValue{Key: key, Value: value, TTL: ttl})
	db.l.Unlock()

	return nil
}
