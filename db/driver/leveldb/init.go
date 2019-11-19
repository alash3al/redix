package leveldb

import (
	"github.com/alash3al/redix/db/driver"
)

const (
	name = "leveldb"
)

func init() {
	driver.Registry["leveldb"] = Driver{}
}
