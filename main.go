package main

import (
	"github.com/alash3al/redix/db"
	"github.com/alash3al/redix/pb"

	_ "github.com/alash3al/goukv/providers/leveldb"
)

func main() {
	db, err := db.Open("leveldb://./redix")
	if err != nil {
		panic(err)
	}

	panic(pb.ListenAndServe(":3035", db))
}
