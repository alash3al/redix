package main

import (
	"flag"
	"log"

	"github.com/alash3al/redix/configparser"
	"github.com/alash3al/redix/redis"
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

	log.Fatal(redis.ListenAndServe(config))
}
