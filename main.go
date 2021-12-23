package main

import (
	"fmt"
	"log"
	"time"

	"github.com/alash3al/redix/internals/datastore/contract"
	"github.com/alash3al/redix/internals/datastore/manager"

	_ "github.com/alash3al/redix/internals/datastore/engines/boltdb"
)

func main() {
	mngr, err := manager.New(&manager.Options{
		DataDir:       "./redixdata",
		DefaultEngine: "boltdb",
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(mngr.Put(0, &contract.PutInput{
		Key:   []byte("key1"),
		Value: []byte("value1"),
	}))

	fmt.Println(mngr.Put(0, &contract.PutInput{
		Key:             []byte("key1"),
		Value:           []byte("value1-overrider"),
		OnlyIfNotExists: true,
		TTL:             time.Second * 5,
	}))

	time.Sleep(1 * time.Second)

	o, err := mngr.Get(0, &contract.GetInput{
		Key: []byte("key1"),
	})

	fmt.Println(err, string(o.Value), "[Will Be Expired After] =>", o.ExpiresAfterSeconds)
}
