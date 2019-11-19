package db

import (
	"fmt"
	"path/filepath"

	"github.com/alash3al/redix/db/driver"
)

// DB represents datastore
type DB struct {
	db     driver.Interface
	name   string
	driver string
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
		db:     driverInstance,
		name:   dbname,
		driver: drivername,
	}

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

// Put puts new document into the storage
func (db DB) Put(key []byte, value []byte) error {
	return db.db.Put(key, value)
}

// Get fetches a document using its primary key
func (db DB) Get(key []byte) ([]byte, error) {
	return db.db.Get(key)
}

// Delete remove a key from the database
func (db DB) Delete(key []byte) error {
	return db.db.Delete(key)
}

// Scan scans the db
func (db DB) Scan(opts driver.ScanOpts) {
	if opts.Filter == nil {
		return
	}

	db.db.Scan(opts)
}

// Close the database
func (db DB) Close() error {
	return db.db.Close()
}
