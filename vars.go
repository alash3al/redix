package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/dgraph-io/badger"

	"github.com/alash3al/color"
)

var (
	flagListenAddr = flag.String("l", "localhost:6380", "the address to listen on")
	flagStorageDir = flag.String("s", "./redix-data", "the storage directory")
	flagVerbose    = flag.Bool("v", false, "verbose or not")
)

var (
	databases *sync.Map
)

var commands = map[string]CommandHandler{
	// strings
	"set":    setCommand,
	"mset":   msetCommand,
	"get":    getCommand,
	"mget":   mgetCommand,
	"del":    delCommand,
	"exists": existsCommands,

	// lists
	"lpush":  lpushCommand,
	"lpushu": lpushuCommand,
	"lrange": lrangeCommand,
	"lrem":   lremCommand,
	"lcount": lcountCommand,

	// hashes
	"hset":    hsetCommand,
	"hget":    hgetCommand,
	"hdel":    hdelCommand,
	"hgetall": hgetallCommand,
	"hmset":   hmsetCommand,
	"hexists": hexistsCommand,
}

func init() {
	flag.Parse()

	if !*flagVerbose {
		logger := logrus.New()
		logger.SetOutput(ioutil.Discard)
		badger.SetLogger(logger)
	}

	databases = new(sync.Map)

	os.MkdirAll(*flagStorageDir, 0744)

	dirs, _ := ioutil.ReadDir(*flagStorageDir)

	for _, f := range dirs {
		if !f.IsDir() {
			continue
		}
		name := filepath.Base(f.Name())
		_, err := selectDB(name)
		if err != nil {
			log.Println(color.RedString(err.Error()))
			continue
		}
	}
}
