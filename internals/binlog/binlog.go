package binlog

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/vmihailenco/msgpack/v5"
)

// BinLog represents a binlog defination
type BinLog struct {
	db      *leveldb.DB
	counter uint64
}

// Open opens the specified path
func Open(path string) (*BinLog, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}

	binlog := &BinLog{
		db:      db,
		counter: 0,
	}

	// FIXME Do we really need this?
	go (func() {
		for {
			atomic.AddUint64(
				&binlog.counter,
				atomic.LoadUint64(&binlog.counter),
			)

			time.Sleep(1 * time.Hour)
		}
	})()

	return binlog, nil
}

// Put inserts the specified entry into our binlog
func (l *BinLog) Put(entry *LogEntry) error {
	timeNs := time.Now().UnixNano()
	id := atomic.AddUint64(&l.counter, 1)

	entry.TimeNS, entry.ID = timeNs, int64(id)

	rawValue, err := msgpack.Marshal(entry)
	if err != nil {
		return err
	}

	return l.db.Put(
		[]byte(entry.Key()),
		rawValue,
		&opt.WriteOptions{
			Sync: true,
		},
	)
}

// ForEach iterate over the binlog using the specified offset and callback fn, if fn retrurns false, means break loop
func (l *BinLog) ForEach(offset []byte, includeOffset bool, fn func(*LogEntry) bool) error {
	iter := l.db.NewIterator(&util.Range{Start: offset}, nil)
	defer iter.Release()

	skippedOffset := false

	for iter.Next() {
		if offset != nil && !includeOffset && !skippedOffset {
			skippedOffset = true
			continue
		}

		var entry LogEntry

		if err := msgpack.Unmarshal(iter.Value(), &entry); err != nil {
			return fmt.Errorf("unable to continue iteration due to: %s", err.Error())
		}

		if !fn(&entry) {
			break
		}
	}

	return iter.Error()
}
