package main

import (
	"fmt"
	"log"

	"github.com/alash3al/redix/internals/datastore/contract"
	_ "github.com/alash3al/redix/internals/datastore/engines/boltdb"
)

func main() {
	e, err := contract.Open("boltdb", "./redixdata/bolt/default-db")
	if err != nil {
		log.Fatal(err)
	}

	if err := e.AddReplica("r-1"); err != nil {
		log.Fatal("ADD REPLICA: ", err)
	}

	if _, err := e.Put(&contract.PutInput{
		Key:   []byte("key-1"),
		Value: []byte("value-1"),
	}); err != nil {
		log.Fatal(err)
	}

	fmt.Println(e.Get(&contract.GetInput{
		Key: []byte("key-1"),
	}))
}
