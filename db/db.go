package db

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/alash3al/goukv"
)

type DB struct {
	kv    goukv.Provider
	queue chan *goukv.Entry
}

type ScanOpts = goukv.ScanOpts

const (
	ttlNoTTL   = -1
	ttlExpired = 0
)

func Open(dsn string) (*DB, error) {
	parts := strings.SplitN(dsn, "://", 2)
	if len(parts) < 1 {
		return nil, errors.New("invalid dsn specified")
	}

	p, err := goukv.Open(parts[0], dsn)
	if err != nil {
		return nil, err
	}

	db, err := &DB{
		kv:    p,
		queue: make(chan *goukv.Entry),
	}, nil

	for i := 0; i < runtime.NumCPU(); i++ {
		go (func() {
			for entry := range db.queue {
				db.handleError(db.commit(entry))
			}
		})()
	}

	return db, err
}

func (db *DB) Put(key, value []byte, ttlDuration string) (err error) {
	var dur time.Duration

	ttlDuration = strings.TrimSpace(ttlDuration)

	if ttlDuration != "" {
		dur, err = time.ParseDuration(ttlDuration)
		if err != nil {
			return
		}
	} else {
		dur = -1
	}

	db.queue <- &goukv.Entry{
		Key:   key,
		Value: value,
		TTL:   dur,
	}

	return nil
}

func (db *DB) PutIfNotExists(key, val []byte, ttl string) error {
	exists, err := db.Has(key)
	if err != nil {
		return err
	}

	if !exists {
		return db.Put(key, val, ttl)
	}

	return nil
}

func (db *DB) Incr(key []byte, delta float64, ttl string) (float64, error) {
	shouldResetValue := false
	ms, err := db.TTL(key)
	if err != nil {
		return 0, err
	}

	if ms != ttlNoTTL && ms == ttlExpired {
		shouldResetValue = true
	} else if ms > 0 {
		ttl = fmt.Sprintf("%dms", ms)
	}

	var floatVal float64

	if !shouldResetValue {
		val, err := db.Get(key)
		if err != nil {
			return 0, err
		}

		floatVal, err = strconv.ParseFloat(string(val), 64)
		if err != nil {
			return 0, err
		}
	}

	floatVal += delta

	_ = db.Put(key, []byte(fmt.Sprintf("%f", floatVal)), ttl)

	return floatVal, nil
}

func (db *DB) Get(key []byte) ([]byte, error) {
	val, err := db.kv.Get(key)
	if err == goukv.ErrKeyExpired {
		_ = db.Put(key, nil, "")
	}

	if err == goukv.ErrKeyNotFound {
		return nil, nil
	}

	return val, err
}

func (db *DB) Has(key []byte) (bool, error) {
	val, err := db.Get(key)
	if err != nil {
		return false, err
	}

	return (val != nil), nil
}

func (db *DB) TTL(key []byte) (int64, error) {
	t, err := db.kv.TTL(key)
	if err != nil {
		return 0, err
	}

	if t == nil {
		return ttlNoTTL, nil
	}

	diff := time.Since(*t).Milliseconds()
	if diff < 0 {
		_ = db.Put(key, nil, "")
		return ttlExpired, nil
	}

	return diff, nil
}

func (db *DB) Scan(opts ScanOpts) error {
	return db.kv.Scan(opts)
}

func (db *DB) Close() error {
	return db.kv.Close()
}

func (db *DB) commit(entry *goukv.Entry) error {
	if entry.Value == nil {
		return db.kv.Delete(entry.Key)
	}
	return db.kv.Put(entry)
}

func (db *DB) handleError(err error) {
	if err == nil {
		return
	}

	fmt.Fprintln(os.Stderr, "[critical_error]:", err.Error())
}
