package main

import (
	"flag"
	"fmt"
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

	if !supportedEngines[*flagEngine] {
		fmt.Println(color.RedString("Invalid strorage engine specified"))
		os.Exit(0)
		return
	}

	color.Cyan(redixBrand)

	databases = new(sync.Map)
	changelog = pubsub.NewBroker()
	webhooks = new(sync.Map)
	websockets = new(sync.Map)

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
