package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	server "github.com/alash3al/redix/internal/server/resp"
	"github.com/alash3al/redix/pkg/db"
	"github.com/tidwall/gjson"

	_ "github.com/alash3al/redix/internal/server/resp/commands"
	_ "github.com/alash3al/redix/pkg/db/providers/badgerdb"
	_ "github.com/alash3al/redix/pkg/db/providers/leveldb"
)

var (
	flagRespListenAddr = flag.String("listen.resp", ":6380", "the local interface address to bind the server to")
	flagDataDriver     = flag.String("storage.driver", "badgerdb", "the storage driver to use")
	flagDataDir        = flag.String("storage.datadir", "./.redix", "the storage data directory")
	flagDriverOpts     = flag.String("storage.opts", "", "the storage engine options")
	flagVerbose        = flag.Bool("verbose", false, "whether to enable verbose mode or not")
)

var (
	parsedDriverOpts gjson.Result
)

func main() {
	flag.Parse()

	*flagDataDir = filepath.Join(*flagDataDir, *flagDataDriver)

	parsedDriverOpts = gjson.Parse(*flagDriverOpts)

	initDBs()

	serverOpts := server.Options{
		Verbose: *flagVerbose,
		Openner: func(dbname string) (*db.DB, error) {
			return db.Open(*flagDataDriver, *flagDataDir, dbname, parsedDriverOpts)
		},
		RESPAddr: *flagRespListenAddr,
	}

	fmt.Println("=> redis server is running on address", *flagRespListenAddr)
	fmt.Printf("=> selected storage driver provider is (%s) with options (%s) \n", *flagDataDriver, *flagDriverOpts)
	fmt.Printf("=> redix store data in (%s) \n", *flagDataDir)

	defer db.CloseAll()

	log.Fatal(server.ListenAndServe(serverOpts))
}

func initDBs() {
	// build the data dir if not
	os.MkdirAll(*flagDataDir, 0755)

	// ping the default db '0'
	if _, err := db.Open(*flagDataDriver, *flagDataDir, "0", parsedDriverOpts); err != nil {
		log.Fatal(err.Error())
	}

	dirs, _ := ioutil.ReadDir(*flagDataDir)

	for _, f := range dirs {
		if !f.IsDir() {
			continue
		}

		name := filepath.Base(f.Name())

		_, err := db.Open(*flagDataDriver, *flagDataDir, name, parsedDriverOpts)
		if err != nil {
			log.Fatal(err.Error())
			continue
		}
	}
}
