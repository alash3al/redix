// Copyright 2018 The Redix Authors. All rights reserved.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.
//
// kvstore is the main interface that any of db engine MUST implement
package kvstore

// DB Interface
type DB interface {
	Incr(k string, by int64) (int64, error)
	Set(k, v string, ttl int) error
	MSet(data map[string]string) error
	Get(k string) (string, error)
	MGet(keys []string) []string
	TTL(key string) int64
	Del(keys []string) error
	Scan(ScannerOpt ScannerOptions) error
	Size() int64
	GC() error
	Close()
}

// ScannerOptions - represents the options for a scanner
type ScannerOptions struct {
	// from where to start
	Offset string

	// whether to include the value of the offset in the result or not
	IncludeOffset bool

	// the prefix that must be exists in each key in the iteration
	Prefix string

	// fetch the values (true) or this is a key only iteration (false)
	FetchValues bool

	// the handler that handles the incoming data
	Handler func(k, v string) bool
}
