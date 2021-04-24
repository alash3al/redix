// Package store provider the contract for each store adapter
package store

import (
	"github.com/alash3al/redix/configparser"
	"github.com/alash3al/redix/redis/context"
)

type Store interface {
	Connect(*configparser.Config) (Store, error)
	Close() error

	IsAuthRequired() bool

	AuthCreate() (string, error)
	AuthReset(token string) (string, error)
	AuthValidate(token string) (bool, error)

	Select(*context.Context, string) error

	// Set(key, subkey []byte, value []byte) error
	// Get(key, subkey []byte) ([]byte, error)
	// SetGet(key, subkey []byte) ([]byte, error)
	// Incr(key, subkey []byte, value float64) (float64, error)
}
