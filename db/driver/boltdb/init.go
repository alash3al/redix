package boltdb

import (
	"github.com/alash3al/redix/db/driver"
)

const (
	name = "boltdb"
)

func init() {
	driver.Registry["boltdb"] = Driver{}
}
