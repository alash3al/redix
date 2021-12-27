package wal

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// RangeOpts represents options passed to the range iterator
type RangeOpts struct {
	Offset             []byte
	IncludeOffsetValue bool
	Limit              int64
}

// Wal represents our Write-Ahead-Log
type Wal struct {
	db      *leveldb.DB
	counter uint64
}

// Open opens the specified path
func Open(path string) (*Wal, error) {
	db, err := leveldb.OpenFile(path, &opt.Options{
		BlockCacher:       opt.NoCacher,
		DisableBlockCache: true,
		Strict:            opt.StrictRecovery,
	})
	if err != nil {
		return nil, err
	}

	wal := &Wal{
		db:      db,
		counter: 0,
	}

	// our counter resetter
	go (func() {
		for {
			atomic.AddUint64(
				&wal.counter,
				atomic.LoadUint64(&wal.counter),
			)

			time.Sleep(1 * time.Hour)
		}
	})()

	return wal, nil
}

// Put inserts the specified entry into our Wal
func (wal *Wal) Write(value []byte) error {
	timeNs := time.Now().UnixNano()
	id := atomic.AddUint64(&wal.counter, 1)
	keyStr := fmt.Sprintf("%d-%d", timeNs, id)
	keyBytes := []byte(keyStr)

	if err := wal.db.Put(keyBytes, value, nil); err != nil {
		return err
	}

	return nil
}

// Read an entry from the wal using the specified offset
// an offset is the key retruned via a Put operation
func (wal *Wal) Read(offset []byte) ([]byte, error) {
	val, err := wal.db.Get(offset, nil)
	if err != nil && err != leveldb.ErrNotFound {
		return nil, err
	}

	return val, nil
}

// Range iterate over the Wal using the specified callback fn and opts, if fn retrurns false, means break loop
func (wal *Wal) Range(fn func([]byte, []byte) bool, opts *RangeOpts) error {
	iter := wal.db.NewIterator(&util.Range{Start: opts.Offset}, nil)
	defer iter.Release()

	skippedOffset := false
	var fetchedCount int = 0

	for iter.Next() {
		if opts.Limit > 0 && fetchedCount >= int(opts.Limit) {
			break
		}

		if opts.Offset != nil && !opts.IncludeOffsetValue && !skippedOffset {
			skippedOffset = true
			continue
		}

		fetchedCount++

		srcValue := iter.Value()
		dstValue := make([]byte, len(srcValue))

		srcKey := iter.Key()
		dstKey := make([]byte, len(srcKey))

		copy(dstValue, srcValue)
		copy(dstKey, srcKey)

		if !fn(dstKey, dstValue) {
			break
		}
	}

	return iter.Error()
}
