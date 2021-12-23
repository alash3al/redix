package boltdb

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alash3al/redix/internals/datastore/contract"
	"github.com/gosimple/slug"
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

	replicas     map[string]*bbolt.DB
	replicasLock *sync.RWMutex

	datadir string
}

// Open opens the specified database file
func (db *DB) Open(dirname string) error {
	if err := os.MkdirAll(dirname, 0755); err != nil {
		return err
	}

	masterDBFilename := filepath.Join(dirname, "master.rxdb")

	bolt, err := bbolt.Open(masterDBFilename, 0666, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}

	if err = bolt.Update(func(tx *bbolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(dataBucketName); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists(expirationsBucketName); err != nil {
			return err
		}
		return err
	}); err != nil {
		return err
	}

	db.DB = bolt
	db.datadir = dirname
	db.replicasLock = new(sync.RWMutex)
	db.replicas = make(map[string]*bbolt.DB)

	repliacsFilesNames, err := filepath.Glob(filepath.Join(dirname, "/*.replica.rxdb"))
	if err != nil {
		return err
	}

	for _, filename := range repliacsFilesNames {
		fileparts := strings.Split(filepath.Base(filename), ".replica.rxdb")
		if len(fileparts) < 1 {
			continue
		}

		replica, err := bbolt.Open(filename, 0666, &bbolt.Options{Timeout: 1 * time.Second})
		if err != nil {
			return err
		}

		db.replicasLock.Lock()
		db.replicas[fileparts[0]] = replica
		db.replicasLock.Unlock()
	}

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
				tx.Bucket(dataBucketName).Put(k, nil)
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

	// TODO move to db.rsync
	err := db.rsync(func(tx *bbolt.Tx) error {
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

// ForEach iterate over each key-value in the datastore
func (db *DB) ForEach(fn contract.IteratorFunc) error {
	stop := errors.New("STOP_ITERATION")

	err := db.View(func(tx *bbolt.Tx) error {
		return tx.Bucket(dataBucketName).ForEach(func(k, v []byte) error {
			if !fn(k, v) {
				return stop
			}

			return nil
		})
	})

	if err == stop {
		return nil
	}

	return err
}

// AddReplica adds a new replica db
func (db *DB) AddReplica(name string) error {
	name = slug.Make(name)
	exists := false

	db.replicasLock.RLock()
	if _, exists = db.replicas[name]; exists {
		exists = true
	}
	db.replicasLock.RUnlock()

	if exists {
		return nil
	}

	replicaFilename := filepath.Join(db.datadir, name+".replica.rxdb")

	fd, err := os.OpenFile(replicaFilename, os.O_CREATE|os.O_TRUNC|os.O_RDWR|os.O_SYNC, 0755)
	if err != nil {
		return err
	}

	defer (func() {
		if fd != nil {
			fd.Close()
		}
	})()

	if err := db.View(func(tx *bbolt.Tx) error {
		_, err := tx.WriteTo(fd)
		return err
	}); err != nil {
		return err
	}

	if err := fd.Close(); err != nil {
		return err
	}

	fd = nil

	replica, err := bbolt.Open(replicaFilename, 0666, &bbolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return err
	}

	db.replicasLock.Lock()
	db.replicas[name] = replica
	defer db.replicasLock.Unlock()

	return nil
}

func (db *DB) rsync(rsyncFn func(*bbolt.Tx) error) error {
	replicas := []*bbolt.DB{db.DB}

	db.replicasLock.RLock()

	for _, replica := range replicas {
		replicas = append(replicas, replica)
	}

	db.replicasLock.RUnlock()

	for _, replica := range replicas {
		if err := replica.Batch(rsyncFn); err != nil {
			return err
		}
	}

	return nil
}
