package boltdb

import (
	"fmt"
	"strconv"
	"time"

	"github.com/alash3al/redix/internals/datastore/contract"
	"go.etcd.io/bbolt"
)

// DB represents the database defeination
type DB struct {
	*bbolt.DB
}

var (
	dataBucketName        = []byte("data")
	expirationsBucketName = []byte("expirations")
)

// Open opens the specified database file
func (db *DB) Open(dsn string) error {
	bolt, err := bbolt.Open(dsn, 0666, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}

	if err = bolt.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(dataBucketName)
		return err
	}); err != nil {
		return err
	}

	db.DB = bolt

	return nil
}

// Get performs the specified Get request
func (db *DB) Get(input *contract.GetInput) (*contract.GetOutput, error) {
	var val []byte
	var expiresAt []byte

	err := db.View(func(tx *bbolt.Tx) error {
		val = tx.Bucket(dataBucketName).Get(input.Key)
		expiresAt = tx.Bucket(expirationsBucketName).Get(input.Key)

		return nil
	})

	if err != nil {
		return nil, err
	}

	expiresAtNano, _ := strconv.ParseInt(string(expiresAt), 10, 64)
	now := time.Now().UnixNano()
	expired := now == expiresAtNano || now < expiresAtNano

	if input.Delete || expired {
		if _, err := db.Delete(&contract.DeleteInput{Key: input.Key}); err != nil {
			return nil, err
		}
	}

	return &contract.GetOutput{
		Value:               val,
		ExpiresAfterSeconds: time.Since(time.Unix(0, expiresAtNano)).Seconds(),
	}, err
}

// Delete performs the specified Delete request
func (db *DB) Delete(input *contract.DeleteInput) (*contract.DeleteOutput, error) {
	err := db.Batch(func(tx *bbolt.Tx) error {
		if err := tx.Bucket(dataBucketName).Delete(input.Key); err != nil {
			return err
		}

		return tx.Bucket(expirationsBucketName).Delete(input.Key)
	})

	return new(contract.DeleteOutput), err
}

// Put performs the specified put request
func (db *DB) Put(input *contract.PutInput) (*contract.PutOutput, error) {
	if input.OnlyIfNotExists {
		getOutput, err := db.Get(&contract.GetInput{Key: input.Key})
		if err != nil {
			return nil, err
		}

		// exists & not expired
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

	return new(contract.PutOutput), err
}
