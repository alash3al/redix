package main

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/alash3al/redix/kvstore/badger"

	"github.com/alash3al/redix/kvstore"
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
	dbpath = dbpath + "-" + engine
	switch strings.ToLower(engine) {
	default:
		return nil, errors.New("unsupported engine: " + engine)
	case "badger":
		return badger.OpenBadger(dbpath)
	}
}
