// Copyright 2018 The Redix Authors. All rights reserved.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.
//
// level is a db engine based on leveldb
package level

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alash3al/redix/kvstore"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// LevelDB - represents a level db implementation
type LevelDB struct {
	leveldb       *leveldb.DB
	countersLocks sync.RWMutex
}

// OpenLevelDB - Opens the specified path
func OpenLevelDB(path string) (*LevelDB, error) {
	ldb, err := leveldb.OpenFile(path, nil)
	if nil != err {
		return nil, err
	}

	db := new(LevelDB)
	db.leveldb = ldb
	db.countersLocks = sync.RWMutex{}

	return db, nil
}

// Size - returns the size of the database (LSM + ValueLog) in bytes
func (db *LevelDB) Size() int64 {
	var stats leveldb.DBStats
	if nil != db.leveldb.Stats(&stats) {
		return -1
	}
	size := int64(0)
	for _, v := range stats.LevelSizes {
		size += v
	}
	return size
}

// GC - runs the garbage collector
func (db *LevelDB) GC() error {
	return db.leveldb.CompactRange(util.Range{})
}

// Incr - increment the key by the specified value
func (db *LevelDB) Incr(k string, by int64) (int64, error) {
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
func (db *LevelDB) Set(k, v string, ttl int) error {
	var expires int64
	if ttl > 0 {
		expires = time.Now().Add(time.Duration(ttl) * time.Millisecond).Unix()
	}
	v = strconv.Itoa(int(expires)) + ";" + v
	return db.leveldb.Put([]byte(k), []byte(v), nil)
}

// MSet - sets multiple key-value pairs
func (db *LevelDB) MSet(data map[string]string) error {
	batch := new(leveldb.Batch)
	for k, v := range data {
		v = "0;" + v
		batch.Put([]byte(k), []byte(v))
	}
	return db.leveldb.Write(batch, nil)
}

// Get - fetches the value of the specified k
func (db *LevelDB) Get(k string) (string, error) {
	var data string

	delete := false

	value, err := db.leveldb.Get([]byte(k), nil)
	if value == nil || err != nil {
		return "", errors.New("key not found")
	}

	parts := strings.SplitN(string(value), ";", 2)
	expires, actual := parts[0], parts[1]
	if exp, _ := strconv.Atoi(expires); exp > 0 && int(time.Now().Unix()) >= exp {
		delete = true
	} else {
		data = actual
	}

	if delete {
		go db.leveldb.Delete([]byte(k), nil)
	}

	return data, err
}

// MGet - fetch multiple values of the specified keys
func (db *LevelDB) MGet(keys []string) (data []string) {
	for _, key := range keys {
		value, err := db.Get(key)
		if err != nil {
			data = append(data, "")
			continue
		}
		data = append(data, strings.SplitN(string(value), ";", 2)[1])
	}

	return data
}

// TTL - returns the time to live of the specified key's value
func (db *LevelDB) TTL(key string) int64 {
	var expires int64

	value, err := db.leveldb.Get([]byte(key), nil)
	if err != nil || value == nil {
		return -2
	}

	parts := strings.SplitN(string(value), ";", 2)
	exp, _ := strconv.Atoi(parts[0])

	if exp == 0 {
		return -1
	}

	if int(time.Now().Unix()) >= exp {
		return -2
	}

	expires = int64(exp)
	now := time.Now().Unix()

	if now >= expires {
		return -2
	}

	return (expires - now)
}

// Del - removes key(s) from the store
func (db *LevelDB) Del(keys []string) error {
	batch := new(leveldb.Batch)
	for _, k := range keys {
		batch.Delete([]byte(k))
	}
	return db.leveldb.Write(batch, nil)
}

// Scan - iterate over the whole store using the handler function
func (db *LevelDB) Scan(scannerOpt kvstore.ScannerOptions) error {
	it := db.leveldb.NewIterator(nil, nil)
	defer it.Release()

	start := func(it iterator.Iterator) {
		it.First()
		if scannerOpt.Offset != "" {
			it.Seek([]byte(scannerOpt.Offset))
			if !scannerOpt.IncludeOffset && it.Valid() {
				it.Next()
			}
		}
	}

	valid := func(it iterator.Iterator) bool {
		if !it.Valid() {
			return false
		}

		if scannerOpt.Prefix != "" && !bytes.HasPrefix(it.Key(), []byte(scannerOpt.Prefix)) {
			return false
		}

		return true
	}

	for start(it); valid(it); it.Next() {
		k, v := it.Key(), ""
		if scannerOpt.FetchValues {
			v = strings.SplitN(string(it.Value()), ";", 2)[1]
		}

		if !scannerOpt.Handler(string(k), string(v)) {
			break
		}
	}

	return it.Error()
}
