package main

import (
	"path/filepath"

	"github.com/dgraph-io/badger"
)

func selectDB(n string) (*DB, error) {
	dbpath := filepath.Join(*flagStorageDir, n)
	dbi, found := databases.Load(n)
	if !found {
		bdb, err := openDB(dbpath)
		if err != nil {
			return nil, err
		}

		databases.Store(n, bdb)
		dbi, _ = databases.Load(n)
	}

	db, _ := dbi.(*badger.DB)

	return &DB{badger: db}, nil
}

func openDB(path string) (*badger.DB, error) {
	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}
	return db, nil
}
