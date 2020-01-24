package badgerdb

import (
	"time"

	"github.com/alash3al/redix/pkg/db/driver"
	"github.com/dgraph-io/badger/v2"
)

// Driver represents a driver
type Driver struct {
	db *badger.DB
}

// Open implements driver.Open
func (drv Driver) Open(dbname string, opts map[string]interface{}) (driver driver.Interface, err error) {
	badgerOpts := badger.DefaultOptions(dbname)
	// badgerOpts.Truncate = true
	// badgerOpts.SyncWrites = false
	// badgerOpts.TableLoadingMode = options.FileIO
	// badgerOpts.ValueLogLoadingMode = options.FileIO
	// badgerOpts.NumMemtables = 1
	// badgerOpts.MaxTableSize = 1 << 20
	// badgerOpts.NumLevelZeroTables = 1
	// badgerOpts.ValueThreshold = 1
	// badgerOpts.KeepL0InMemory = false

	db, err := badger.Open(badgerOpts)

	if err != nil {
		return driver, err
	}

	go (func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			for {
				err := db.RunValueLogGC(0.5)
				if err != nil {
					break
				}
			}
		}
	})()

	return Driver{
		db: db,
	}, nil
}

// Put implements driver.Put
func (drv Driver) Put(k, v []byte) error {
	return drv.db.Update(func(txn *badger.Txn) error {
		return txn.Set(k, v)
	})
}

// Batch perform multi put operation, empty value means *delete*
func (drv Driver) Batch(pairs []driver.Pair) error {
	batch := drv.db.NewWriteBatch()

	for _, pair := range pairs {
		if pair.Value == nil {
			batch.Delete(pair.Key)
		} else {
			batch.Set(pair.Key, pair.Value)
		}
	}

	return batch.Flush()
}

// Get implements driver.Get
func (drv Driver) Get(k []byte) ([]byte, error) {
	var data []byte
	err := drv.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(k)
		if err != nil {
			return err
		}

		d, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		data = d

		return err
	})

	return data, err
}

// Has implements driver.Has
func (drv Driver) Has(k []byte) (bool, error) {
	data, err := drv.Get(k)

	if err == badger.ErrKeyNotFound {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return data != nil, nil
}

// Delete implements driver.Delete
func (drv Driver) Delete(k []byte) error {
	return drv.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(k)
	})
}

// Close implements driver.Close
func (drv Driver) Close() error {
	return drv.db.Close()
}

// Scan implements driver.Scan
func (drv Driver) Scan(opts driver.ScanOpts) {
	if opts.Scanner == nil {
		return
	}

	txn := drv.db.NewTransaction(false)
	defer txn.Commit()

	iterOpts := badger.DefaultIteratorOptions
	iterOpts.Reverse = opts.ReverseScan

	iter := txn.NewIterator(iterOpts)
	defer iter.Close()

	valid := func() bool {
		if opts.Prefix != nil {
			return iter.ValidForPrefix(opts.Prefix)
		}

		return iter.Valid()
	}

	rewind := func() {
		if opts.Offset != nil {
			iter.Seek(opts.Offset)
		} else {
			iter.Rewind()
		}
	}

	for rewind(); valid(); iter.Next() {
		item := iter.Item()

		key := item.KeyCopy(nil)
		val, err := item.ValueCopy(nil)
		if err != nil {
			break
		}

		if !opts.Scanner(key, val) {
			break
		}
	}
}
