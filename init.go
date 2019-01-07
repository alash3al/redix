// Copyright 2018 The Redix Authors. All rights reserved.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/alash3al/go-color"
	"github.com/alash3al/go-pubsub"
	"github.com/bwmarrin/snowflake"
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

	databases = new(sync.Map)
	changelog = pubsub.NewBroker()
	webhooks = new(sync.Map)
	websockets = new(sync.Map)
	engineOptions = (func() url.Values {
		opts, _ := url.ParseQuery(*flagEngineOpions)
		return opts
	})()
	flagStorageDir = (func() *string {
		ret := filepath.Join(*flagStorageDir, *flagEngine)
		return &ret
	})()

	snowflakenode, err := snowflake.NewNode(1)
	if err != nil {
		fmt.Println(color.RedString(err.Error()))
		os.Exit(0)
		return
	}

	snowflakeGenerator = snowflakenode

	initDBs()
}

// initDBs - initialize databases from the disk for faster access
func initDBs() {
	os.MkdirAll(*flagStorageDir, 0755)

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
