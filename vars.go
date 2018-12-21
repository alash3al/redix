package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/alash3al/color"
	"github.com/alash3al/go-pubsub"

	"github.com/dgraph-io/badger"
	"github.com/sirupsen/logrus"
)

var (
	flagListenAddr = flag.String("l", "localhost:6380", "the address to listen on")
	flagStorageDir = flag.String("s", "./redix-data", "the storage directory")
	flagEngine     = flag.String("e", "badger", "the storage engine to be used, available (badger)")
	flagWorkers    = flag.Int("w", runtime.NumCPU()*4, "the default workers number")
	flagVerbose    = flag.Bool("v", false, "verbose or not")
)

var (
	databases *sync.Map
	changelog *pubsub.Broker
	webhooks  *sync.Map
)

var (
	commands = map[string]CommandHandler{
		// strings
		"set":    setCommand,
		"mset":   msetCommand,
		"get":    getCommand,
		"mget":   mgetCommand,
		"del":    delCommand,
		"exists": existsCommand,
		"incr":   incrCommand,

		// lists
		"lpush":      lpushCommand,
		"lpushu":     lpushuCommand,
		"lrange":     lrangeCommand,
		"lrem":       lremCommand,
		"lcount":     lcountCommand,
		"lsum":       lsumCommand,
		"lavg":       lavgCommand,
		"lmin":       lminCommand,
		"lmax":       lmaxCommand,
		"lsrch":      lsearchCommand,
		"lsrchcount": lsearchcountCommand,

		// hashes
		"hset":    hsetCommand,
		"hget":    hgetCommand,
		"hdel":    hdelCommand,
		"hgetall": hgetallCommand,
		"hmset":   hmsetCommand,
		"hexists": hexistsCommand,
		"hincr":   hincrCommand,

		// pubsub
		"publish":    publishCommand,
		"subscribe":  subscribeCommand,
		"webhookset": webhooksetCommand,
		"webhookdel": webhookdelCommand,

		// utils
		"encode":  encodeCommand,
		"uuidv4":  uuid4Command,
		"uniqid":  uniqidCommand,
		"randstr": randstrCommand,
		"randint": randintCommand,
		"time":    timeCommand,
	}

	defaultPubSubAllTopic = "*"
)

func init() {
	flag.Parse()

	runtime.GOMAXPROCS(*flagWorkers)

	if !*flagVerbose {
		logger := logrus.New()
		logger.SetOutput(ioutil.Discard)
		badger.SetLogger(logger)
	}

	databases = new(sync.Map)
	changelog = pubsub.NewBroker()
	webhooks = new(sync.Map)

	*flagStorageDir = filepath.Join(*flagStorageDir, *flagEngine)

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
