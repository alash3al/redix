package boltdb

import (
	"bytes"

	"github.com/alash3al/redix/pkg/db/driver"

	bbolt "go.etcd.io/bbolt"
)

// Driver - represents a badger db implementation
type Driver struct {
	bolt *bbolt.DB
}

// Open - Opens the specified path
func (db Driver) Open(path string, opts map[string]interface{}) (driver.Interface, error) {
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

	return Driver{
		bolt: bdb,
	}, nil
}

// Close ...
func (db Driver) Close() error {
	return db.bolt.Close()
}

// Size - returns the size of the database (LSM + ValueLog) in bytes
func (db Driver) Size() int64 {
	var size int64

	db.bolt.View(func(txn *bbolt.Tx) error {
		size = txn.Size()
		return nil
	})

	return size
}

// GC - runs the garbage collector
func (db Driver) GC() error {
	return nil
}

// Put - sets a key with the specified value and optional ttl
func (db Driver) Put(k, v []byte) error {
	return db.bolt.Update(func(txn *bbolt.Tx) (err error) {
		return txn.Bucket([]byte("default")).Put(k, v)
	})
}

// Get - fetches the value of the specified k
func (db Driver) Get(k []byte) ([]byte, error) {
	var data []byte

	err := db.bolt.View(func(txn *bbolt.Tx) error {
		data = txn.Bucket([]byte("default")).Get(k)

		return nil
	})

	return data, err
}

// Has implements driver.Has
func (db Driver) Has(k []byte) (bool, error) {
	data, err := db.Get(k)

	if err != nil {
		return false, err
	}

	return data != nil, nil
}

// Delete - removes key(s) from the store
func (db Driver) Delete(key []byte) error {
	return db.bolt.Update(func(txn *bbolt.Tx) error {
		return txn.Bucket([]byte("default")).Delete(key)
	})
}

// Scan - iterate over the whole store using the handler function
func (db Driver) Scan(scannerOpt driver.ScanOpts) {
	db.bolt.View(func(txn *bbolt.Tx) error {
		var k, v []byte

		it := txn.Bucket([]byte("default")).Cursor()

		start := func(it *bbolt.Cursor) {
			if scannerOpt.Offset == nil {
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

			if scannerOpt.Prefix != nil && !bytes.HasPrefix(k, scannerOpt.Prefix) {
				return false
			}

			return true
		}

		for start(it); valid(it); k, v = it.Next() {
			if !scannerOpt.Scanner(k, v) {
				break
			}
		}

		return nil
	})
}
