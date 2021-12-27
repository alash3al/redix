package main

import (
	"fmt"
	"log"
	"os"

	"github.com/alash3al/redix/internals/config"
	"github.com/alash3al/redix/internals/datastore/engines/boltdb"
	"github.com/alash3al/redix/internals/manager"
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

	mngr, err := manager.New(&manager.Options{
		DataDir:       cfg.DataDir,
		DefaultEngine: boltdb.EngineName,
		InstanceRole:  cfg.InstanceRole,
		MasterDSN:     cfg.MasterDSN,
	})

	if err != nil {
		log.Fatal("unable to initialize the database manager due to: ", err.Error())
	}

	fmt.Println("=> started the redis server on addr", cfg.InstanceRespListenAddr)

	log.Fatal(
		"unable to start the redis server due to: ",
		redis.ListenAndServe(cfg.InstanceRespListenAddr, mngr),
	)
}
