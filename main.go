package main

import (
	"fmt"
	"log"
	"os"

	"github.com/alash3al/redix/internals/config"
	"github.com/alash3al/redix/internals/datastore/engines/boltdb"
	"github.com/alash3al/redix/internals/http"
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
		DataDir:           cfg.DataDir,
		DefaultEngine:     boltdb.EngineName,
		InstanceRole:      cfg.InstanceRole,
		MasterRESPDSN:     cfg.MasterRESPDSN,
		MasterHTTPBaseURL: cfg.MasterHTTPBaseURL,
		MaxWalSize:        cfg.MaxWalSize,
		RESPListenAddr:    cfg.InstanceRESPListenAddr,
	})

	if err != nil {
		log.Fatal("unable to initialize the database manager due to: ", err.Error())
	}

	go func() {
		fmt.Println("=> started the redis server on addr", cfg.InstanceRESPListenAddr)
		log.Fatal(
			"unable to start the redis server due to: ",
			redis.ListenAndServe(cfg.InstanceRESPListenAddr, mngr),
		)
	}()

	fmt.Println("=> started the http server on addr", cfg.InstanceHTTPListenAddr)
	log.Fatal(
		"unable to start the http server due to: ",
		http.ListenAndServe(cfg.InstanceHTTPListenAddr, mngr),
	)
}
