// Copyright 2018 The Redix Authors. All rights reserved.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.
//
// null is a db engine based on `/dev/null` style
package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/alash3al/redix/kvstore"
)

// SQLite - represents a SQLite db implementation
type SQLite struct {
	db       *sql.DB
	filename string
	sync.RWMutex
}

// OpenSQLite - Opens the specified path
func OpenSQLite(dbpath string) (*SQLite, error) {
	db, err := sql.Open("sqlite3", dbpath+"?cache=shared&_journal=wal")
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS data (
			_id INTEGER PRIMARY KEY,
			_key TEXT NOT NULL UNIQUE,
			_value TEXT NOT NULL,
			_expires INTEGER
		);
	`); err != nil {
		return nil, err
	}

	lite := new(SQLite)
	lite.filename = dbpath
	lite.db = db

	lite.db.Exec("PRAGMA synchronous = OFF")
	lite.db.Exec("PRAGMA journal_mode = OFF")
	lite.db.Exec(fmt.Sprintf("PRAGMA page_size = %d", 20*1024*1024))

	lite.GC()

	return lite, nil
}

// Size - returns the size of the database in bytes
func (db *SQLite) Size() int64 {
	stat, err := os.Stat(db.filename)
	if err != nil {
		return 0
	}
	return stat.Size()
}

// GC - runs the garbage collector
func (db *SQLite) GC() error {
	_, err := db.db.Exec("VACUUM")
	return err
}

// Incr - increment the key by the specified value
func (db *SQLite) Incr(k string, by int64) (int64, error) {
	db.Lock()
	defer db.Unlock()

	val, err := db.Get(k)
	if err != nil {
		val = ""
	}

	valInt, _ := strconv.ParseInt(val, 10, 64)
	valInt += by

	err = db.Set(k, strconv.FormatInt(valInt, 10), -1)
	if err != nil {
		return 0, err
	}

	return valInt, nil
}

// Set - sets a key with the specified value and optional ttl
func (db *SQLite) Set(k, v string, ttl int) error {
	var expires int64

	if ttl < 1 {
		expires = 0
	} else {
		expires = time.Now().Add(time.Duration(ttl) * time.Millisecond).Unix()
	}

	_, err := db.db.Exec("INSERT OR REPLACE INTO data(_key, _value, _expires) VALUES(?, ?, ?)", k, v, expires)
	return err
}

// MSet - sets multiple key-value pairs
func (db *SQLite) MSet(data map[string]string) error {
	args := []interface{}{}
	values := []string{}
	for k, v := range data {
		values = append(values, "(?, ?, ?)")
		args = append(args, k, v, 0)
	}
	_, err := db.db.Exec("INSERT OR REPLACE INTO data(_key, _value, _expires) VALUES"+strings.Join(values, ", "), args...)
	return err
}

// Get - fetches the value of the specified k
func (db *SQLite) Get(k string) (string, error) {
	var value string
	var expires int64

	if err := db.db.QueryRow("SELECT _value, _expires FROM data WHERE _key = ?", k).Scan(&value, &expires); err != nil {
		return "", err
	}

	now := time.Now().Unix()

	if now >= expires && expires > 0 {
		db.db.Exec("DELETE FROM data WHERE _key = ?", k)
		return "", errors.New("key not found")
	}

	return value, nil
}

// MGet - fetch multiple values of the specified keys
func (db *SQLite) MGet(keys []string) (data []string) {
	for _, key := range keys {
		v, _ := db.Get(key)
		data = append(data, v)
	}
	return data
}

// TTL - returns the time to live of the specified key's value
func (db *SQLite) TTL(key string) int64 {
	var value string
	var expires int64

	if err := db.db.QueryRow("SELECT _value, _expires FROM data WHERE _key = ?", key).Scan(&value, &expires); err != nil {
		return -2
	}

	if expires < 1 {
		return -1
	}

	now := time.Now().Unix()

	if now >= expires {
		return -2
	}

	return (expires - now)
}

// Del - removes key(s) from the store
func (db *SQLite) Del(keys []string) error {
	for _, k := range keys {
		db.db.Exec("DELETE FROM data WHERE _key = ?", k)
	}
	return nil
}

// Scan - iterate over the whole store using the handler function
func (db *SQLite) Scan(scannerOpt kvstore.ScannerOptions) error {
	sql := "SELECT _key, _value FROM data"
	wheres := []string{}
	args := []interface{}{}

	if scannerOpt.Offset != "" {
		op := ">"
		if scannerOpt.IncludeOffset {
			op = ">="
		}
		wheres = append(wheres, "_id "+(op)+" (SELECT _id FROM data WHERE _key LIKE ?)")
		args = append(args, scannerOpt.Offset+"%")
	}

	if scannerOpt.Prefix != "" {
		wheres = append(wheres, "_key LIKE ?")
		args = append(args, scannerOpt.Prefix+"%")
	}

	if len(wheres) > 0 {
		sql += " WHERE (" + strings.Join(wheres, ") AND (") + ")"
	}

	rows, err := db.db.Query(sql, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var k, v string

		if err := rows.Scan(&k, &v); err != nil {
			return err
		}

		if next := scannerOpt.Handler(k, v); !next {
			break
		}
	}

	return nil
}
