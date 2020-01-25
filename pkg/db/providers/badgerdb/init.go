package badgerdb

import (
	"github.com/alash3al/redix/pkg/db/driver"
)

const (
	name = "badgerdb"
)

func init() {
	driver.Register("badgerdb", Driver{})
}
