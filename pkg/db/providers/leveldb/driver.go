package leveldb

import (
	"fmt"

	"github.com/alash3al/redix/pkg/db/driver"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/tidwall/gjson"
)

// Driver represents a driver
type Driver struct {
	db         *leveldb.DB
	syncWrites bool
}

// Open implements driver.Open
func (drv Driver) Open(dbname string, opts gjson.Result) (driver driver.IDriver, err error) {
	syncWrites := !opts.Get("sync_writes").Exists() || opts.Get("sync_writes").Bool()
	o := &opt.Options{
		Filter:         filter.NewBloomFilter(10),
		ErrorIfMissing: false,
		Compression:    9,
		NoSync:         syncWrites,
	}

	db, err := leveldb.OpenFile(dbname, o)
	if err != nil {
		return driver, err
	}

	return Driver{
		db:         db,
		syncWrites: syncWrites,
	}, nil
}

// Put implements driver.Put
func (drv Driver) Put(e driver.Entry) error {
	return drv.db.Put(e.Key, EntryToValue(e).Bytes(), &opt.WriteOptions{
		Sync: drv.syncWrites,
	})
}

// Batch perform multi put operation, empty value means *delete*
func (drv Driver) Batch(entries []driver.Entry) error {
	batch := new(leveldb.Batch)

	for _, entry := range entries {
		fmt.Println(entry)
		if entry.Value == nil {
			batch.Delete(entry.Key)
		} else {
			batch.Put(entry.Key, EntryToValue(entry).Bytes())
		}
	}

	return drv.db.Write(batch, &opt.WriteOptions{
		Sync: drv.syncWrites,
	})
}

// Get implements driver.Get
func (drv Driver) Get(k []byte) ([]byte, error) {
	b, err := drv.db.Get(k, nil)
	if err == leveldb.ErrNotFound {
		return nil, nil
	}

	val := BytesToValue(b)

	if val.Expires != nil && val.IsExpired() {
		return nil, driver.ErrKeyExpired
	}

	return val.Value, err
}

// Has implements driver.Has
func (drv Driver) Has(k []byte) (bool, error) {
	b, err := drv.Get(k)
	return b != nil, err
}

// Delete implements driver.Delete
func (drv Driver) Delete(k []byte) error {
	return drv.db.Delete(k, &opt.WriteOptions{
		Sync: drv.syncWrites,
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

	var iter iterator.Iterator
	var next func() bool

	if opts.Prefix != nil {
		iter = drv.db.NewIterator(util.BytesPrefix(opts.Prefix), nil)
	} else {
		iter = drv.db.NewIterator(nil, nil)
	}

	if opts.ReverseScan {
		next = iter.Prev
	} else {
		next = iter.Next
	}

	if opts.Offset != nil {
		iter.Seek(opts.Offset)
	}

	if opts.ReverseScan && opts.Offset == nil && opts.Prefix == nil {
		iter.Last()
	}

	if opts.Offset != nil && !opts.IncludeOffset {
		next()
	}

	defer iter.Release()
	for next() {
		if err := iter.Error(); err != nil {
			break
		}

		if !iter.Valid() {
			break
		}

		_k, _v := iter.Key(), iter.Value()

		if _k == nil {
			break
		}

		newK := make([]byte, len(_k))
		newV := make([]byte, len(_v))

		copy(newK, _k)
		copy(newV, _v)

		decodedValue := BytesToValue(newV)
		if decodedValue.IsExpired() {
			continue
		}

		if !opts.Scanner(newK, newV) {
			break
		}
	}
}
