package db

import (
	"fmt"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/alash3al/redix/db/driver"
)

var (
	countersPrefix  = []byte("internal::counters::")
	expirablePrefix = []byte("internal::expirable::")
)

// DB represents datastore
type DB struct {
	db       driver.Interface
	name     string
	driver   string
	writes   atomic.Value
	counters atomic.Value
	tmp      chan driver.KeyValue
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
		db:       driverInstance,
		name:     dbname,
		driver:   drivername,
		writes:   atomic.Value{},
		counters: atomic.Value{},
	}

	db.writes.Store([]driver.KeyValue{})

	go (func() {
		tk := time.NewTicker(100 * time.Millisecond)
		defer tk.Stop()

		for range tk.C {
			snapshot := db.writes.Load()
			db.writes.Store([]driver.KeyValue{})
			pairs := snapshot.([]driver.KeyValue)

			for _, pair := range pairs {
				if pair.Value == nil {
					db.delete(pair.Key)
				} else {
					db.put(pair.Key, pair.Value, pair.TTL)
				}
			}
		}
	})()

	counters := map[string]int64{}
	db.Scan(driver.ScanOpts{
		Prefix:        countersPrefix,
		IncludeOffset: true,
		Scanner: func(k, v []byte) bool {
			k = k[len(countersPrefix)+1:]

			val, _ := strconv.ParseInt(string(v), 10, 64)
			counters[string(k)] = val

			return true
		},
	})
	db.counters.Store(counters)

	databases.Store(dbkey, db)

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

// PutAsync puts new document into the storage [async]
func (db *DB) PutAsync(key []byte, value []byte, ttl int) error {
	var newVals []driver.KeyValue

	pair := driver.KeyValue{Key: key, Value: value, TTL: ttl}
	oldVals := db.writes.Load().([]driver.KeyValue)

	copy(newVals, oldVals)

	newVals = append(newVals, pair)

	db.writes.Store(newVals)

	return nil
}

// Put puts new document into the storage [sync]
func (db DB) Put(key []byte, value []byte, ttl int) error {
	return db.put(key, value, ttl)
}

// // Batch a bulk pairs writer
// func (db DB) Batch(pairs []driver.KeyValue) error {
// 	return db.db.Batch(pairs)
// }

// Incr increments a key
func (db *DB) Incr(key []byte, delta int) (int64, error) {
	oldCounters := db.counters.Load().(map[string]int64)
	newCounters := map[string]int64{}

	for k, v := range oldCounters {
		newCounters[k] = v
	}

	key = append(countersPrefix, key...)

	newCounters[string(key)] += int64(delta)
	newVal := newCounters[string(key)]

	db.counters.Store(newCounters)

	return newVal, db.PutAsync(key, []byte(strconv.FormatInt(newVal, 10)), 0)
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
		db.DeleteAsync(key)
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

// DeleteAsync remove a key from the database [async]
func (db DB) DeleteAsync(key []byte) error {
	pair := driver.KeyValue{Key: key, Value: nil}

	snapshot := db.writes.Load()
	vals := snapshot.([]driver.KeyValue)
	vals = append(vals, pair)
	db.writes.Store(vals)

	return nil
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

// Delete remove a key from the database [async]
func (db DB) delete(key []byte) error {
	return db.db.Delete(key)
}

// put puts new document into the storage [sync]
func (db DB) put(key []byte, value []byte, ttl int) error {
	if ttl > 0 {
		expires := time.Now().Add(time.Duration(ttl) * time.Millisecond).UnixNano()
		db.db.Put(append(expirablePrefix, key...), []byte(strconv.FormatInt(expires, 10)))
	}

	return db.db.Put(key, value)
}
