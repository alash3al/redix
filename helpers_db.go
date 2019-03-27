// Copyright 2018 The Redix Authors. All rights reserved.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.
package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/alash3al/redix/kvstore"
	"github.com/alash3al/redix/kvstore/badgerdb"
	"github.com/alash3al/redix/kvstore/boltdb"
	"github.com/alash3al/redix/kvstore/leveldb"
	"github.com/alash3al/redix/kvstore/null"
)

// selectDB - load/fetches the requested db
func selectDB(n string) (db kvstore.DB, err error) {
	dbpath := filepath.Join(*flagStorageDir, n)
	dbi, found := databases.Load(n)
	if !found {
		db, err = openDB(*flagEngine, dbpath)
		if err != nil {
			return nil, err
		}
		databases.Store(n, db)
	} else {
		db, _ = dbi.(kvstore.DB)
	}

	return db, nil
}

// openDB - initialize a db in the specified path and engine
func openDB(engine, dbpath string) (kvstore.DB, error) {
	switch strings.ToLower(engine) {
	default:
		return nil, errors.New("unsupported engine: " + engine)
	case "badgerdb":
		return badgerdb.OpenBadger(dbpath)
	case "boltdb":
		return boltdb.OpenBolt(dbpath)
	case "leveldb":
		return leveldb.OpenLevelDB(dbpath)
	case "null":
		return null.OpenNull()
	}
}

// flushDB clear the specified database
func flushDB(n string) {
	dbpath := filepath.Join(*flagStorageDir, n)
	os.RemoveAll(dbpath)
	databases.Delete(n)

	selectDB(n)
}

// flushall clear all databases
func flushall() {
	os.RemoveAll(*flagStorageDir)
	os.MkdirAll(*flagStorageDir, 0755)
}

// returns a unique string
func getUniqueString() string {
	return snowflakeGenerator.Generate().String()
}

// returns a unique int
func getUniqueInt() int64 {
	return snowflakeGenerator.Generate().Int64()
}
