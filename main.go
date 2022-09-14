//go:build linux || darwin

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/alash3al/redix/internals/config"
	"github.com/alash3al/redix/internals/datastore/contract"
	_ "github.com/alash3al/redix/internals/datastore/engines/postgresql"
	"github.com/alash3al/redix/internals/redis"
)

var (
	cfg *config.Config
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("you must specify the configuration file as an argument")
	}

	var err error

	fmt.Println("=> loading the configs ...")

	cfg, err = config.Unmarshal(os.Args[1])
	if err != nil {
		log.Fatal("unable to load the config file due to: ", err.Error())
	}

	db, err := contract.Open(cfg.Engine.Driver, cfg.Engine.DSN)
	if err != nil {
		log.Fatal("failed to open database connection due to: ", err.Error())
	}

	redis.ListenAndServe(cfg, db)
}
