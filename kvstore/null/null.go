// Copyright 2018 The Redix Authors. All rights reserved.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.
//
// null is a db engine based on `/dev/null` style
package null

import (
	"github.com/alash3al/redix/kvstore"
)

// Null - represents a Null db implementation
type Null struct{}

// OpenNull - Opens the specified path
func OpenNull() (*Null, error) {
	return new(Null), nil
}

// Size - returns the size of the database in bytes
func (ldb *Null) Size() int64 {
	return 0
}

// GC - runs the garbage collector
func (ldb *Null) GC() error {
	return nil
}

// Incr - increment the key by the specified value
func (ldb *Null) Incr(k string, by int64) (int64, error) {
	return 1, nil
}

// Set - sets a key with the specified value and optional ttl
func (ldb *Null) Set(k, v string, ttl int) error {
	return nil
}

// MSet - sets multiple key-value pairs
func (ldb *Null) MSet(data map[string]string) error {
	return nil
}

// Get - fetches the value of the specified k
func (ldb *Null) Get(k string) (string, error) {
	return "", nil
}

// MGet - fetch multiple values of the specified keys
func (ldb *Null) MGet(keys []string) []string {
	return nil
}

// TTL - returns the time to live of the specified key's value
func (ldb *Null) TTL(key string) int64 {
	return -2
}

// Del - removes key(s) from the store
func (ldb *Null) Del(keys []string) error {
	return nil
}

// Scan - iterate over the whole store using the handler function
func (ldb *Null) Scan(scannerOpt kvstore.ScannerOptions) error {
	return nil
}
