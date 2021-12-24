package main

import (
	"github.com/alash3al/redix/internals/redis"

	_ "github.com/alash3al/redix/internals/datastore/engines/boltdb"
)

func main() {
	redis.ListenAndServe(":4000")
}
