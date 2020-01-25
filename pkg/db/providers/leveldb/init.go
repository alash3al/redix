package leveldb

import (
	"github.com/alash3al/redix/pkg/db/driver"
)

const (
	name = "leveldb"
)

func init() {
	driver.Register("leveldb", Driver{})
}
