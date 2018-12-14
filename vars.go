package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

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
	"ping":    pingCommand,
	"set":     setCommand,
	"mset":    msetCommand,
	"get":     getCommand,
	"mget":    mgetCommand,
	"del":     delCommand,
	"scan":    scanCommand,
	"append":  appendCommand,
	"mappend": mappendCommand,
	"hset":    hsetCommand,
	"hdel":    hdelCommand,
	"hgetall": hgetallCommand,
	"hmset":   hmsetCommand,
}

func init() {
	flag.Parse()

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
