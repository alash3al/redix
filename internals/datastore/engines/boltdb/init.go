package boltdb

import "github.com/alash3al/redix/internals/datastore/contract"

// Engine level consts
const (
	EngineName = "boltdb"
)

func init() {
	contract.Register(EngineName, &DB{})
}
