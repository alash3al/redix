package boltdb

import (
	"fmt"
	"strconv"
	"time"

	"github.com/alash3al/redix/internals/datastore/contract"
	"go.etcd.io/bbolt"
)

// Engine level vars
var (
	dataBucketName        = []byte("data")
	expirationsBucketName = []byte("expirations")
)

// DB represents the database defeination
type DB struct {
	*bbolt.DB
	statemachine contract.Engine
}

// Open opens the specified database file
func (db *DB) Open(dsn string) error {
	bolt, err := bbolt.Open(dsn, 0666, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}

	if err = bolt.Update(func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(dataBucketName); err != nil {
			return err
		}
		_, err := tx.CreateBucketIfNotExists(expirationsBucketName)
		return err
	}); err != nil {
		return err
	}

	db.DB = bolt

	go (func() {
		expiredKeys := [][]byte{}

		// FIXME handle any error happends during view?
		db.View(func(tx *bbolt.Tx) error {
			return tx.Bucket(expirationsBucketName).ForEach(func(k, v []byte) error {
				expiresAtNano, _ := strconv.ParseInt(string(v), 10, 64)
				now := time.Now().UnixNano()

				if now >= expiresAtNano {
					expiredKeys = append(expiredKeys, k)
				}

				return nil
			})
		})

		// FIXME handle any error happends during batch update?
		db.Batch(func(tx *bbolt.Tx) error {
			for _, k := range expiredKeys {
				tx.Bucket(dataBucketName).Delete(k)
			}

			return nil
		})

		time.Sleep(10 * time.Minute)
	})()

	return nil
}

// Get performs the specified Get request
func (db *DB) Get(input *contract.GetInput) (*contract.GetOutput, error) {
	var val []byte
	var expiresAtRaw []byte
	var expiresAtParsed time.Time
	var expired bool
	var expiresAfterSeconds float64

	now := time.Now()

	err := db.View(func(tx *bbolt.Tx) error {
		val = tx.Bucket(dataBucketName).Get(input.Key)
		expiresAtRaw = tx.Bucket(expirationsBucketName).Get(input.Key)

		return nil
	})

	if err != nil {
		return nil, err
	}

	if expiresAtRaw != nil {
		expiresAtNano, _ := strconv.ParseInt(string(expiresAtRaw), 10, 64)
		expired = now.UnixNano() >= expiresAtNano
		expiresAtParsed = time.Unix(0, expiresAtNano)

		if !expired {
			expiresAfterSeconds = expiresAtParsed.Sub(now).Seconds()
		}
	}

	if expired {
		return new(contract.GetOutput), nil
	}

	return &contract.GetOutput{
		Value:               val,
		ExpiresAfterSeconds: expiresAfterSeconds,
	}, err
}

// Delete performs the specified Delete request
func (db *DB) Delete(input *contract.DeleteInput) (*contract.DeleteOutput, error) {
	err := db.Batch(func(tx *bbolt.Tx) error {
		if err := tx.Bucket(dataBucketName).Delete(input.Key); err != nil {
			return err
		}

		if err := tx.Bucket(expirationsBucketName).Delete(input.Key); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return new(contract.DeleteOutput), nil
}

// Put performs the specified put request
func (db *DB) Put(input *contract.PutInput) (*contract.PutOutput, error) {
	if input.OnlyIfNotExists {
		getOutput, err := db.Get(&contract.GetInput{Key: input.Key})
		if err != nil {
			return nil, err
		}

		if getOutput.Value != nil && getOutput.ExpiresAfterSeconds >= 0 {
			return new(contract.PutOutput), nil
		}
	}

	err := db.Batch(func(tx *bbolt.Tx) error {
		if err := tx.Bucket(dataBucketName).Put(
			input.Key,
			input.Value,
		); err != nil {
			return err
		}

		if input.TTL > 0 {
			return tx.Bucket(expirationsBucketName).Put(
				input.Key,
				[]byte(fmt.Sprintf("%v", time.Now().Add(input.TTL).UnixNano())),
			)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return new(contract.PutOutput), nil
}
