package badgerdb

import (
	"time"

	"github.com/tidwall/gjson"

	"github.com/alash3al/redix/pkg/db/driver"
	"github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/options"
)

// Driver represents a driver
type Driver struct {
	db *badger.DB
}

// Open implements driver.Open
func (drv Driver) Open(dbname string, opts gjson.Result) (driver driver.IDriver, err error) {
	badgerOpts := badger.LSMOnlyOptions(dbname)
	badgerOpts.Compression = options.Snappy
	badgerOpts.SyncWrites = !opts.Get("sync_writes").Exists() || opts.Get("sync_writes").Bool()
	badgerOpts.Logger = nil

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
func (drv Driver) Put(entry driver.Entry) error {
	return drv.db.Update(func(txn *badger.Txn) error {
		if entry.TTL > 0 {
			badgerEntry := badger.NewEntry(entry.Key, entry.Value)
			badgerEntry.WithTTL(time.Duration(entry.TTL) * time.Second)
			return txn.SetEntry(badgerEntry)
		}

		return txn.Set(entry.Key, entry.Value)
	})
}

// Batch perform multi put operation, empty value means *delete*
func (drv Driver) Batch(entries []driver.Entry) error {
	batch := drv.db.NewWriteBatch()
	defer batch.Cancel()

	for _, entry := range entries {
		var err error
		if entry.Value == nil {
			err = batch.Delete(entry.Key)
		} else {
			if entry.TTL > 0 {
				badgerEntry := badger.NewEntry(entry.Key, entry.Value)
				badgerEntry.WithTTL(time.Duration(entry.TTL) * time.Second)

				err = batch.SetEntry(badgerEntry)
			} else {
				err = batch.Set(entry.Key, entry.Value)
			}
		}

		if err != nil {
			batch.Cancel()
			return err
		}
	}

	return batch.Flush()
}

// Get implements driver.Get
func (drv Driver) Get(k []byte) ([]byte, error) {
	var data []byte
	err := drv.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(k)
		if err == badger.ErrKeyNotFound {
			data = nil
			return nil
		}
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

	if len(opts.Prefix) > 0 {
		iterOpts.Prefix = opts.Prefix
	}

	iter := txn.NewIterator(iterOpts)
	defer iter.Close()

	if opts.Offset != nil {
		iter.Seek(opts.Offset)
	} else {
		iter.Rewind()
	}

	for ; iter.Valid(); iter.Next() {
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
