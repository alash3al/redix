package main

import (
	"flag"
	"runtime"
	"sync"

	"github.com/alash3al/go-pubsub"
)

var (
	flagListenAddr = flag.String("l", "localhost:6380", "the address to listen on")
	flagStorageDir = flag.String("s", "./redix-data", "the storage directory")
	flagEngine     = flag.String("e", "badger", "the storage engine to be used, available (badger)")
	flagGCInterval = flag.Int("gc", 30, "databse GC interval in seconds")
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
		// internals
		"dbsize": dbsizeCommand,

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
