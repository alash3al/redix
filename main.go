package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/alash3al/redix/configparser"
	"github.com/alash3al/redix/driver"
	"github.com/alash3al/redix/engine"
)

var (
	flagConfigFilename = flag.String("config", "./redix.hcl", "the configuration filename")
)

func main() {
	flag.Parse()

	config, err := configparser.Parse(*flagConfigFilename)
	if err != nil {
		log.Fatal(err.Error())
	}

	fmt.Println(config)

	e, _ := engine.New()

	e.Put(&driver.RawValueEntry{
		Entry: driver.Entry{
			Key: "path1/key1",
		},
		Value: "value1",
	})

	e.Put(&driver.RawValueEntry{
		Entry: driver.Entry{
			Key: "path2/key1",
		},
		Value: "value1",
	})

	e.Walk(func(rve *driver.RawValueEntry) bool {
		fmt.Printf("%v\t", rve)
		return false
	})

	fmt.Println(e.Get("path1/key1"))
}
