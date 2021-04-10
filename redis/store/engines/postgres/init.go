package postgres

import (
	"github.com/alash3al/redix/configparser"
	"github.com/alash3al/redix/redis/store/engines"
)

const engineName = configparser.DatabaseEngine("postgres")

func init() {
	engines.RegisterStorageEngine(engineName, &Store{})
}
