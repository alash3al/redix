// Package store provider the contract for each store adapter
package store

import (
	"errors"

	"github.com/alash3al/redix/internals/configparser"
)

var (
	ErrCommandNotImplemented = errors.New("command not implemented")
)

type Store interface {
	Connect(*configparser.Config) (Store, error)
	Close() error

	AuthCreate() (string, error)
	AuthReset(token string) (string, error)
	AuthValidate(token string) (bool, error)

	Select(token string, db int) (int, error)

	Exec(command string, args ...string) (interface{}, error)
}
