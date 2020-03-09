package db

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/alash3al/goukv"
)

type DB struct {
	store        goukv.Provider
	singleWrites chan *goukv.Entry
	batchWrites  chan []*goukv.Entry
	stop         bool
}

func newdb(path string, opts *Options) (*DB, error) {
	store, err := goukv.Open(string(opts.Provider), map[string]interface{}{
		"path": path,
	})
	if err != nil {
		return nil, err
	}

	db := &DB{
		store:        store,
		singleWrites: make(chan *goukv.Entry, opts.WriteQeueSize),
		batchWrites:  make(chan []*goukv.Entry, opts.WriteQeueSize),
	}

	go db.flushBatchWrites()
	go db.flushSingleWrites()

	return db, nil
}

func (db *DB) Put(entry *Entry) {
	if db.stop {
		return
	}

	db.singleWrites <- entry
}

func (db *DB) Batch(entries []*Entry) {
	if db.stop {
		return
	}

	db.batchWrites <- entries
}

func (db *DB) Get(k []byte) ([]byte, error) {
	val, err := db.store.Get(k)
	if err == goukv.ErrKeyExpired {
		entry := Entry{Key: k}
		db.Put(&entry)
		return nil, nil
	}

	return val, err
}

func (db *DB) TTL(k []byte) (*time.Time, error) {
	return db.store.TTL(k)
}

func (db *DB) Incr(k []byte, delta float64, ttl time.Duration) (*float64, error) {
	oldValBin, err := db.Get(k)
	if err != nil {
		return nil, err
	}

	if ttl < 1 {
		oldTTL, err := db.TTL(k)
		if err != nil {
			return nil, err
		}
		if oldTTL != nil {
			ttl = oldTTL.Sub(time.Now())
		}
	}

	oldValFloat, _ := strconv.ParseFloat(string(oldValBin), 64)

	if delta == 0 {
		delta = 1
	}

	oldValFloat += delta

	entry := Entry{
		Key:   k,
		Value: []byte(strconv.FormatFloat(oldValFloat, 'f', -1, 64)),
		TTL:   ttl,
	}

	db.Put(&entry)

	return &oldValFloat, nil
}

func (db *DB) Close() {
	db.stop = true

	for len(db.batchWrites) > 0 {
	}
	for len(db.singleWrites) > 0 {
	}

	if err := db.store.Close(); err != nil {
		if _, err := fmt.Fprintln(os.Stderr, "#close:", err.Error()); err != nil {
			panic("#close: couldn't write an error to the stderr, here is the error")
		}
	}
}

func (db *DB) flushSingleWrites() {
	for entry := range db.singleWrites {
		if entry.Value == nil {
			if err := db.store.Delete(entry.Key); err != nil {
				if _, err := fmt.Fprintln(os.Stderr, "#flushSingleWrites:", err.Error()); err != nil {
					panic("#flushSingleWrites: couldn't write an error to the stderr, here is the error")
				}
			}
			continue
		}

		if err := db.store.Put(entry); err != nil {
			if _, err := fmt.Fprintln(os.Stderr, "#flushSingleWrites:", err.Error()); err != nil {
				panic("#flushSingleWrites: couldn't write an error to the stderr, here is the error")
			}
		}
	}
}

func (db *DB) flushBatchWrites() {
	for entries := range db.batchWrites {
		if err := db.store.Batch(entries); err != nil {
			if _, err := fmt.Fprintln(os.Stderr, "#flushBatchWrites:", err.Error()); err != nil {
				panic("#flushBatchWrites: couldn't write an error to the stderr, here is the error")
			}
		}
	}
}
