package store

import "github.com/alash3al/redix/configparser"

type Connector interface {
	Connect(*configparser.Config) (Store, error)
	Close() error
}
