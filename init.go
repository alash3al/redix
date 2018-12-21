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
	pubsub "github.com/alash3al/go-pubsub"
	"github.com/dgraph-io/badger"
	"github.com/sirupsen/logrus"
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
