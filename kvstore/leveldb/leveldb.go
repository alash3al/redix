// Copyright 2018 The Redix Authors. All rights reserved.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.
//
// leveldb is a db engine based on leveldb
package leveldb

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

// LevelDB - represents a leveldb db implementation
type LevelDB struct {
	db *leveldb.DB
	sync.RWMutex
}

// OpenLevelDB - Opens the specified path
func OpenLevelDB(path string) (*LevelDB, error) {
	db, err := leveldb.OpenFile(path, nil)

	if err != nil {
		return nil, err
	}

	ldb := new(LevelDB)
	ldb.db = db

	return ldb, nil
}

// Size - returns the size of the database in bytes
func (ldb *LevelDB) Size() int64 {
	ldb.RLock()
	defer ldb.RUnlock()

	// TODO need to find efficiency method stat dbsize
	return int64(0)
}

// GC - runs the garbage collector
func (ldb *LevelDB) GC() error {
	return nil
}

// Incr - increment the key by the specified value
func (ldb *LevelDB) Incr(k string, by int64) (int64, error) {
	ldb.Lock()
	defer ldb.Unlock()

	val, err := ldb.get(k)
	if err != nil {
		val = ""
	}

	valFloat, _ := strconv.ParseInt(val, 10, 64)
	valFloat += by

	err = ldb.set(k, fmt.Sprintf("%d", valFloat), -1)
	if err != nil {
		return 0, err
	}

	return valFloat, nil
}

func (ldb *LevelDB) set(k, v string, ttl int) error {
	var expires int64
	if ttl > 0 {
		expires = time.Now().Add(time.Duration(ttl) * time.Millisecond).Unix()
	}
	v = strconv.Itoa(int(expires)) + ";" + v
	return ldb.db.Put([]byte(k), []byte(v), nil)
}

// Set - sets a key with the specified value and optional ttl
func (ldb *LevelDB) Set(k, v string, ttl int) error {
	ldb.Lock()
	defer ldb.Unlock()

	return ldb.set(k, v, ttl)
}

// MSet - sets multiple key-value pairs
func (ldb *LevelDB) MSet(data map[string]string) error {
	ldb.Lock()
	defer ldb.Unlock()

	batch := new(leveldb.Batch)
	for k, v := range data {
		v = "0;" + v
		batch.Put([]byte(k), []byte(v))
	}

	return ldb.db.Write(batch, nil)
}

func (ldb *LevelDB) get(k string) (string, error) {
	var data string
	var err error

	delete := false

	item, err := ldb.db.Get([]byte(k), nil)
	if err != nil {
		return "", err
	}

	parts := strings.SplitN(string(item), ";", 2)
	expires, actual := parts[0], parts[1]

	if exp, _ := strconv.Atoi(expires); exp > 0 && int(time.Now().Unix()) >= exp {
		delete = true
		err = errors.New("key not found")
	} else {
		data = actual
	}

	if delete {
		ldb.db.Delete([]byte(k), nil)
		return data, errors.New("key not found")
	}

	return data, nil
}

// Get - fetches the value of the specified k
func (ldb *LevelDB) Get(k string) (string, error) {
	ldb.RLock()
	defer ldb.RUnlock()

	return ldb.get(k)
}

// MGet - fetch multiple values of the specified keys
func (ldb *LevelDB) MGet(keys []string) (data []string) {
	ldb.RLock()
	defer ldb.RUnlock()

	for _, key := range keys {
		val, err := ldb.get(key)
		if err != nil {
			data = append(data, "")
			continue
		}
		data = append(data, val)
	}
	return data
}

// TTL - returns the time to live of the specified key's value
func (ldb *LevelDB) TTL(key string) int64 {
	ldb.RLock()
	defer ldb.RUnlock()

	item, err := ldb.db.Get([]byte(key), nil)
	if err != nil {
		return -2
	}

	parts := strings.SplitN(string(item), ";", 2)
	exp, _ := strconv.Atoi(parts[0])
	if exp == 0 {
		return -1
	}

	if int(time.Now().Unix()) >= exp {
		return -2
	}

	return int64(exp)
}

// Del - removes key(s) from the store
func (ldb *LevelDB) Del(keys []string) error {
	ldb.Lock()
	defer ldb.Unlock()

	batch := new(leveldb.Batch)
	for _, key := range keys {
		batch.Delete([]byte(key))
	}

	return ldb.db.Write(batch, nil)
}

// Scan - iterate over the whole store using the handler function
func (ldb *LevelDB) Scan(scannerOpt kvstore.ScannerOptions) error {
	ldb.RLock()
	defer ldb.RUnlock()

	var iter iterator.Iterator

	if scannerOpt.Offset == "" {
		iter = ldb.db.NewIterator(nil, nil)
	} else {
		iter = ldb.db.NewIterator(&util.Range{Start: []byte(scannerOpt.Offset)}, nil)
		if !scannerOpt.IncludeOffset {
			iter.Next()
		}
	}

	valid := func(k []byte) bool {
		if k == nil {
			return false
		}

		if scannerOpt.Prefix != "" && !bytes.HasPrefix(k, []byte(scannerOpt.Prefix)) {
			return false
		}

		return true
	}

	for iter.Next() {
		key := iter.Key()
		val := strings.SplitN(string(iter.Value()), ";", 2)[1]
		if valid(key) && !scannerOpt.Handler(string(key), string(val)) {
			break
		}
	}
	iter.Release()

	return iter.Error()
}
