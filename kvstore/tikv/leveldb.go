// Copyright 2018 The Redix Authors. All rights reserved.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.
//
// tikv is a db engine based on tikv
package tikv

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pingcap/tidb/store/tikv"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/alash3al/redix/kvstore"
)

// TiKV - represents a TiKV implementation
type TiKV struct {
	// db *store
	sync.RWMutex
}

// OpenLevelDB - Opens the specified path
func OpenLevelDB(path string) (*TiKV, error) {
	driver := tikv.Driver{}
	store, err := driver.Open(path)

	return ldb, nil
}

// Size - returns the size of the database in bytes
func (ldb *TiKV) Size() int64 {
	return 0
}

// GC - runs the garbage collector
func (ldb *TiKV) GC() error {
	return nil
}

// Incr - increment the key by the specified value
func (ldb *TiKV) Incr(k string, by int64) (int64, error) {
	ldb.Lock()
	defer ldb.Unlock()

	val, err := ldb.get(k)
	if err != nil {
		val = ""
	}

	valInt, _ := strconv.ParseInt(val, 10, 64)
	valInt += by

	err = ldb.set(k, fmt.Sprintf("%d", valInt), -1)
	if err != nil {
		return 0, err
	}

	return valInt, nil
}

func (ldb *TiKV) set(k, v string, ttl int) error {
	var expires int64
	if ttl > 0 {
		expires = time.Now().Add(time.Duration(ttl) * time.Millisecond).Unix()
	}
	v = strconv.Itoa(int(expires)) + ";" + v
	return ldb.db.Put([]byte(k), []byte(v), nil)
}

// Set - sets a key with the specified value and optional ttl
func (ldb *TiKV) Set(k, v string, ttl int) error {
	return ldb.set(k, v, ttl)
}

// MSet - sets multiple key-value pairs
func (ldb *TiKV) MSet(data map[string]string) error {
	batch := new(leveldb.Batch)
	for k, v := range data {
		v = "0;" + v
		batch.Put([]byte(k), []byte(v))
	}
	return ldb.db.Write(batch, nil)
}

func (ldb *TiKV) get(k string) (string, error) {
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
func (ldb *TiKV) Get(k string) (string, error) {
	return ldb.get(k)
}

// MGet - fetch multiple values of the specified keys
func (ldb *TiKV) MGet(keys []string) (data []string) {
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
func (ldb *TiKV) TTL(key string) int64 {
	item, err := ldb.db.Get([]byte(key), nil)
	if err != nil {
		return -2
	}

	parts := strings.SplitN(string(item), ";", 2)
	exp, _ := strconv.Atoi(parts[0])
	if exp == 0 {
		return -1
	}

	now := time.Now().Unix()
	if now >= int64(exp) {
		return -2
	}

	return int64(exp) - now
}

// Del - removes key(s) from the store
func (ldb *TiKV) Del(keys []string) error {
	batch := new(leveldb.Batch)
	for _, key := range keys {
		batch.Delete([]byte(key))
	}
	return ldb.db.Write(batch, nil)
}

// Scan - iterate over the whole store using the handler function
func (ldb *TiKV) Scan(scannerOpt kvstore.ScannerOptions) error {
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
