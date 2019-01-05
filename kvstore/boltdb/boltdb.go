// Copyright 2018 The Redix Authors. All rights reserved.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.
//
// bolt is a db engine based on boltdb
package boltdb

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alash3al/redix/kvstore"
	bbolt "go.etcd.io/bbolt"
)

// BoltDB - represents a badger db implementation
type BoltDB struct {
	bolt          *bbolt.DB
	countersLocks sync.RWMutex
}

// OpenBolt - Opens the specified path
func OpenBolt(path string) (*BoltDB, error) {
	bdb, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	if err := bdb.Update(func(txn *bbolt.Tx) error {
		_, err := txn.CreateBucketIfNotExists([]byte("default"))
		return err
	}); err != nil {
		return nil, err
	}

	db := new(BoltDB)
	db.bolt = bdb
	db.countersLocks = sync.RWMutex{}

	return db, nil
}

// Size - returns the size of the database (LSM + ValueLog) in bytes
func (db *BoltDB) Size() int64 {
	var size int64

	db.bolt.View(func(txn *bbolt.Tx) error {
		size = txn.Size()
		return nil
	})

	return size
}

// GC - runs the garbage collector
func (db *BoltDB) GC() error {
	return nil
}

// Incr - increment the key by the specified value
func (db *BoltDB) Incr(k string, by int64) (int64, error) {
	db.countersLocks.Lock()
	defer db.countersLocks.Unlock()

	val, err := db.Get(k)
	if err != nil {
		val = ""
	}

	valFloat, _ := strconv.ParseInt(val, 10, 64)
	valFloat += by

	err = db.Set(k, fmt.Sprintf("%d", valFloat), -1)
	if err != nil {
		return 0, err
	}

	return valFloat, nil
}

// Set - sets a key with the specified value and optional ttl
func (db *BoltDB) Set(k, v string, ttl int) error {
	return db.bolt.Update(func(txn *bbolt.Tx) (err error) {
		var expires int64
		if ttl > 0 {
			expires = time.Now().Add(time.Duration(ttl) * time.Millisecond).Unix()
		}
		v := strconv.Itoa(int(expires)) + ";" + v
		return txn.Bucket([]byte("default")).Put([]byte(k), []byte(v))
	})
}

// MSet - sets multiple key-value pairs
func (db *BoltDB) MSet(data map[string]string) error {
	return db.bolt.Update(func(txn *bbolt.Tx) (err error) {
		b := txn.Bucket([]byte("default"))
		for k, v := range data {
			v = "0;" + v
			b.Put([]byte(k), []byte(v))
		}
		return nil
	})
}

// Get - fetches the value of the specified k
func (db *BoltDB) Get(k string) (string, error) {
	var data string

	delete := false

	err := db.bolt.View(func(txn *bbolt.Tx) error {
		b := txn.Bucket([]byte("default"))
		value := b.Get([]byte(k))
		if value == nil {
			return errors.New("key not found")
		}

		parts := strings.SplitN(string(value), ";", 2)
		expires, actual := parts[0], parts[1]
		if exp, _ := strconv.Atoi(expires); exp > 0 && int(time.Now().Unix()) >= exp {
			delete = true
			return errors.New("key not found")
		}

		data = actual

		return nil
	})

	if delete {
		go db.bolt.Update(func(txn *bbolt.Tx) error {
			txn.Bucket([]byte("default")).Delete([]byte(k))
			return nil
		})
	}

	return data, err
}

// MGet - fetch multiple values of the specified keys
func (db *BoltDB) MGet(keys []string) (data []string) {
	db.bolt.View(func(txn *bbolt.Tx) error {
		b := txn.Bucket([]byte("default"))
		for _, key := range keys {
			value := b.Get([]byte(key))
			if value == nil {
				data = append(data, "")
				continue
			}
			data = append(data, strings.SplitN(string(value), ";", 2)[1])
		}
		return nil
	})

	return data
}

// TTL - returns the time to live of the specified key's value
func (db *BoltDB) TTL(key string) int64 {
	var expires int64

	db.bolt.View(func(txn *bbolt.Tx) error {
		b := txn.Bucket([]byte("default"))
		value := b.Get([]byte(key))
		if value == nil {
			expires = -2
			return nil
		}

		parts := strings.SplitN(string(value), ";", 2)
		exp, _ := strconv.Atoi(parts[0])

		if exp == 0 {
			expires = -1
			return nil
		}

		if int(time.Now().Unix()) >= exp {
			expires = -2
		}

		expires = int64(exp)
		return nil
	})

	if expires == -2 {
		return -2
	}

	if expires == -1 {
		return -1
	}

	now := time.Now().Unix()

	if now >= expires {
		return -2
	}

	return (expires - now)
}

// Del - removes key(s) from the store
func (db *BoltDB) Del(keys []string) error {
	return db.bolt.Update(func(txn *bbolt.Tx) error {
		b := txn.Bucket([]byte("default"))
		for _, key := range keys {
			b.Delete([]byte(key))
		}

		return nil
	})
}

// Scan - iterate over the whole store using the handler function
func (db *BoltDB) Scan(scannerOpt kvstore.ScannerOptions) error {
	return db.bolt.View(func(txn *bbolt.Tx) error {
		var k, v []byte

		it := txn.Bucket([]byte("default")).Cursor()

		start := func(it *bbolt.Cursor) {
			if scannerOpt.Offset == "" {
				k, v = it.First()
			} else {
				k, v = it.Seek([]byte(scannerOpt.Offset))
				if !scannerOpt.IncludeOffset && k != nil {
					k, v = it.Next()
				}
			}
		}

		valid := func(it *bbolt.Cursor) bool {
			if k == nil {
				return false
			}

			if scannerOpt.Prefix != "" && !bytes.HasPrefix(k, []byte(scannerOpt.Prefix)) {
				return false
			}

			return true
		}

		for start(it); valid(it); k, v = it.Next() {
			val := strings.SplitN(string(v), ";", 2)[1]
			if !scannerOpt.Handler(string(k), val) {
				break
			}
		}

		return nil
	})
}
