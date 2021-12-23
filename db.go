package main

import (
	"sync"
	"time"

	"go.etcd.io/bbolt"
)

var (
	dataBucketName = []byte("data")
	ttlsBucketName = []byte("ttls")
)

type DB struct {
	*bbolt.DB
	mutex *sync.RWMutex
}

type PutInput struct {
	Key   []byte
	Value []byte
}

type PutOutput struct{}

type GetInput struct {
	Key []byte
}

type GetOutput struct {
	Value []byte
}

type DeleteInput struct {
	Key []byte
}

type DeleteOutput struct{}

func NewDB() (*DB, error) {
	db, err := bbolt.Open("./redixdata/redix.db", 0666, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(dataBucketName)
		return err
	})

	if err != nil {
		return nil, err
	}

	return &DB{DB: db, mutex: new(sync.RWMutex)}, nil
}

func (db *DB) Put(input *PutInput) (*PutOutput, error) {
	err := db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(dataBucketName).Put(
			input.Key,
			input.Value,
		)
	})

	return new(PutOutput), err
}

func (db *DB) Get(input *GetInput) (*GetOutput, error) {
	var val []byte

	err := db.View(func(tx *bbolt.Tx) error {
		val = tx.Bucket(dataBucketName).Get(input.Key)
		return nil
	})

	return &GetOutput{Value: val}, err
}

func (db *DB) Del(input *DeleteInput) (*DeleteOutput, error) {
	err := db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket(dataBucketName).Delete(input.Key)
	})

	return new(DeleteOutput), err
}
