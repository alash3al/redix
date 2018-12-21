package main

import (
	"flag"
	"runtime"
	"sync"

	"github.com/alash3al/go-pubsub"
)

var (
	flagRESPListenAddr = flag.String("resp-addr", "localhost:6380", "the address of resp server")
	flagHTTPListenAddr = flag.String("http-addr", "localhost:7090", "the address of the http server")
	flagStorageDir     = flag.String("storage", "./redix-data", "the storage directory")
	flagEngine         = flag.String("engine", "badger", "the storage engine to be used, available (badger)")
	flagWorkers        = flag.Int("workers", runtime.NumCPU()*4, "the default workers number")
	flagVerbose        = flag.Bool("verbose", false, "verbose or not")
)

var (
	databases  *sync.Map
	changelog  *pubsub.Broker
	webhooks   *sync.Map
	websockets *sync.Map
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
		"publish":        publishCommand,
		"subscribe":      subscribeCommand,
		"webhookset":     webhooksetCommand,
		"webhookdel":     webhookdelCommand,
		"websocketopen":  websocketopenCommand,
		"websocketclose": websocketcloseCommand,

		// utils
		"encode":  encodeCommand,
		"uuidv4":  uuid4Command,
		"uniqid":  uniqidCommand,
		"randstr": randstrCommand,
		"randint": randintCommand,
		"time":    timeCommand,
		"dbsize":  dbsizeCommand,
		"gc":      gcCommand,
	}

	defaultPubSubAllTopic = "*"
)
