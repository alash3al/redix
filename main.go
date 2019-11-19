package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/alash3al/redix/db"
	"github.com/alash3al/redix/server"

	_ "github.com/alash3al/redix/db/driver/leveldb"
	_ "github.com/alash3al/redix/server/handlers"
)

var (
	flagRespListenAddr = flag.String("listen.resp", ":6380", "the local interface address to bind the server to")
	flagDataDriver     = flag.String("storage.driver", "leveldb", "the storage driver to use")
	flagDataDir        = flag.String("storage.datadir", "./.redix", "the storage data directory")
	flagDriverOpts     = flag.String("storage.opts", "", "the storage engine options")
)

var (
	parsedDriverOpts map[string]interface{}
)

func main() {
	flag.Parse()

	if *flagDriverOpts != "" {
		if err := json.Unmarshal([]byte(*flagDriverOpts), &parsedDriverOpts); err != nil {
			log.Fatal("You must specified a valid json string in the storage options flag")
		}
	}

	initDBs()

	serverOpts := server.Options{
		DriverName: *flagDataDriver,
		DriverOpts: parsedDriverOpts,
		RESPAddr:   *flagRespListenAddr,
	}

	fmt.Println("=> redis server is running on address", *flagRespListenAddr)
	fmt.Printf("=> selected storage driver is (%s) with options (%s) \n", *flagDataDriver, *flagDriverOpts)
	fmt.Printf("=> redix store data in (%s) \n", *flagDataDir)

	defer db.CloseAll()

	log.Fatal(server.ListenAndServe(serverOpts))
}

func initDBs() {
	// build the data dir if not
	os.MkdirAll(*flagDataDir, 0755)

	// ping the default db '0'
	if _, err := db.Open(*flagDataDriver, filepath.Join(*flagDataDir, "0"), parsedDriverOpts); err != nil {
		log.Fatal(err.Error())
	}

	dirs, _ := ioutil.ReadDir(*flagDataDir)

	for _, f := range dirs {
		if !f.IsDir() {
			continue
		}

		name := filepath.Base(f.Name())

		_, err := db.Open(*flagDataDriver, filepath.Join(*flagDataDir, name), parsedDriverOpts)
		if err != nil {
			log.Fatal(err.Error())
			continue
		}
	}
}
