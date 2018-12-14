package main

import (
	"regexp"
	"strings"
	"time"

	"github.com/dgraph-io/badger"
)

// DB - represents a DB structure
type DB struct {
	badger *badger.DB
}

// Set - sets a key with the specified value and optional ttl
func (db *DB) Set(k, v string, ttl int) error {
	return db.badger.Update(func(txn *badger.Txn) (err error) {
		if ttl < 1 {
			err = txn.Set([]byte(k), []byte(v))
		} else {
			err = txn.SetWithTTL([]byte(k), []byte(v), time.Duration(ttl)*time.Millisecond)
		}

		return err
	})
}

// MSet - sets multiple key-value pairs
func (db *DB) MSet(data map[string]string) error {
	return db.badger.Update(func(txn *badger.Txn) (err error) {
		for k, v := range data {
			txn.Set([]byte(k), []byte(v))
		}
		return nil
	})
}

// Get - fetches the value of the specified k
func (db *DB) Get(k string) (string, error) {
	var data string

	err := db.badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(k))
		if err != nil {
			return err
		}

		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}

		data = string(val)

		return nil
	})

	return data, err
}

// Del - removes key(s) from the store
func (db *DB) Del(keys []string) error {
	return db.badger.Update(func(txn *badger.Txn) error {
		for _, key := range keys {
			txn.Delete([]byte(key))
		}

		return nil
	})
}

// Scan - iterate over the whole store
func (db *DB) Scan(offset string, keyOnly bool, size int) (result []string, err error) {
	err = db.badger.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = !keyOnly
		it := txn.NewIterator(opts)
		defer it.Close()

		if offset == "" {
			for it.Rewind(); it.Valid() && (size < 1 || len(result) <= size); it.Next() {
				item := it.Item()
				k := item.KeyCopy(nil)

				result = append(result, string(k))

				if !keyOnly {
					v, _ := item.ValueCopy(nil)
					result = append(result, string(v))
				}
			}
		} else if !strings.Contains(offset, "*") && !strings.HasSuffix(offset, "%") {
			for it.Seek([]byte(offset)); it.Valid() && (size < 1 || len(result) <= size); it.Next() {
				item := it.Item()
				k := item.KeyCopy(nil)

				result = append(result, string(k))

				if !keyOnly {
					v, _ := item.ValueCopy(nil)
					result = append(result, string(v))
				}
			}
		} else if strings.HasSuffix(offset, "%") {
			offset = strings.TrimSuffix(offset, "%")
			for it.Seek([]byte(offset)); it.ValidForPrefix([]byte(offset)) && (size < 1 || len(result) <= size); it.Next() {
				item := it.Item()
				k := item.KeyCopy(nil)

				result = append(result, string(k))

				if !keyOnly {
					v, _ := item.ValueCopy(nil)
					result = append(result, string(v))
				}
			}
		} else if strings.Contains(offset, "*") {
			re, err := regexp.Compile(offset)
			if err != nil {
				return err
			}
			for it.Rewind(); it.Valid() && (size < 1 || len(result) <= size); it.Next() {
				item := it.Item()
				k := item.KeyCopy(nil)

				if !re.Match(k) {
					continue
				}

				result = append(result, string(k))

				if !keyOnly {
					v, _ := item.ValueCopy(nil)
					result = append(result, string(v))
				}
			}
		}

		return nil
	})

	return result, err
}

// MGet - fetch multiple values of the specified keys
func (db *DB) MGet(keys []string) (data []string) {
	db.badger.View(func(txn *badger.Txn) error {
		for _, key := range keys {
			item, err := txn.Get([]byte(key))
			if err != nil {
				data = append(data, "")
				continue
			}
			val, err := item.ValueCopy(nil)
			if err != nil {
				data = append(data, "")
				continue
			}
			data = append(data, string(val))
		}
		return nil
	})

	return data
}
