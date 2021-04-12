package postgres

import (
	"github.com/alash3al/redix/configparser"
	"github.com/alash3al/redix/redis/store/engines"
)

const DriverPostgres = configparser.DatabaseEngine("postgres")

func init() {
	engines.RegisterStorageEngine(DriverPostgres, &Store{})
}
