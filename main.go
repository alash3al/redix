package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/alash3al/redix/internals/db"
	"github.com/alash3al/redix/internals/resp"
	_ "github.com/alash3al/redix/internals/resp/commands"
)

var (
	flagRespListenAddr = flag.String("listen.resp", ":6380", "the local interface address to bind the server to")
	flagDataDir        = flag.String("storage.datadir", "./.redix", "the storage data directory")
	flagWriteQueueSize = flag.Int("storage.queue.size", 0, "the write queue size, <= 0 means disable queue and write to the underlying storage directly else, write into the memory first, then flush to the disk")
	flagVerbose        = flag.Bool("verbose", false, "whether to enable verbose mode or not")
)

var (
	manager *db.Manager
)

const (
	defaultDriver = db.LevelDBProvder
)

func main() {
	flag.Parse()

	manager = db.NewManager(*flagDataDir, db.Options{
		WriteQeueSize: *flagWriteQueueSize,
		Provider:      defaultDriver,
	})

	manager.OpenDB("0")

	serverOpts := resp.Options{
		Verbose: *flagVerbose,
		Openner: func(dbname string) (*db.DB, error) {
			return manager.OpenDB(dbname)
		},
		RESPAddr: *flagRespListenAddr,
	}

	fmt.Println("=> redis server is running on address", *flagRespListenAddr)
	fmt.Printf("=> redix store data in (%s) \n", *flagDataDir)

	defer manager.CloseAll()

	log.Fatal(resp.ListenAndServe(serverOpts))
}
