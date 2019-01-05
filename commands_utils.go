// Copyright 2018 The Redix Authors. All rights reserved.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.
package main

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	mathRand "math/rand"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/satori/go.uuid"
)

// uuid4Command - UUID4
func uuid4Command(c Context) {
	id, err := uuid.NewV4()
	if err != nil {
		c.WriteError(err.Error())
		return
	}
	c.WriteString(id.String())
}

// uniqidCommand - UNIQID
func uniqidCommand(c Context) {
	c.WriteString(getUniqueString())
}

// randstrCommand - randstr [<size>, default size is 10]
func randstrCommand(c Context) {
	var size int

	if len(c.args) < 1 {
		size = 10
	} else {
		size, _ = strconv.Atoi(c.args[0])
	}

	if size < 1 {
		size = 10
	}

	b := make([]byte, size)

	if _, err := rand.Read(b); err != nil {
		c.WriteError(err.Error())
		return
	}

	c.WriteString(hex.EncodeToString(b))
}

// randintCommand - RANDINT [<max>, default max is 10]
func randintCommand(c Context) {
	if len(c.args) < 2 {
		c.WriteError("RANDINT command must have at least 2 arguments: IRAND <min> <max>")
		return
	}

	min, _ := strconv.Atoi(c.args[0])
	max, _ := strconv.Atoi(c.args[1])

	i := max - min

	c.WriteInt64(mathRand.Int63n(int64(i)) + int64(min))
}

// timeCommand - TIME
func timeCommand(c Context) {
	now := time.Now()

	c.WriteArray(6)

	c.WriteBulkString("utc")
	c.WriteBulkString(now.UTC().String())

	c.WriteBulkString("seconds")
	c.WriteInt64(now.Unix())

	c.WriteBulkString("nanoseconds")
	c.WriteInt64(now.UnixNano())
}

// encodeCommand - Encode <method> <payload>
func encodeCommand(c Context) {
	methods := map[string]func(string) string{
		"md5": func(s string) string {
			d := md5.Sum([]byte(s))
			return hex.EncodeToString(d[:])
		},
		"sha1": func(s string) string {
			d := sha1.Sum([]byte(s))
			return hex.EncodeToString(d[:])
		},
		"sha256": func(s string) string {
			d := sha256.Sum256([]byte(s))
			return hex.EncodeToString(d[:])
		},
		"sha512": func(s string) string {
			d := sha512.Sum512([]byte(s))
			return hex.EncodeToString(d[:])
		},
		"hex": func(s string) string {
			return hex.EncodeToString([]byte(s))
		},
	}

	if len(c.args) < 2 {
		c.WriteError("ENCODE command requires at least 2 arguments: ENCODE <method> <payload>")
		return
	}

	method, payload := strings.ToLower(c.args[0]), c.args[1]
	if methods[method] == nil {
		c.WriteError("unknown encoding method")
		return
	}

	c.WriteString((methods[method])(payload))
}

// dbsizeCommand - DBSIZE
func dbsizeCommand(c Context) {
	c.WriteInt64(c.db.Size())
}

// gcCommand - GC
func gcCommand(c Context) {
	runtime.GC()
	debug.FreeOSMemory()
	if err := c.db.GC(); err != nil {
		c.WriteError(err.Error())
		return
	}
	c.WriteInt(1)
}

// infoCommand - INFO
func infoCommand(c Context) {
	info := map[string]string{
		"version":            redixVersion,
		"database":           *flagEngine,
		"database_size":      strconv.Itoa(int(c.db.Size())),
		"database_directory": *flagStorageDir,
		"redis_port":         *flagRESPListenAddr,
		"http_port":          *flagHTTPListenAddr,
		"workers":            strconv.Itoa(*flagWorkers),
	}

	c.WriteArray(len(info))
	for k, v := range info {
		c.WriteBulkString(k + " : " + v)
	}
}

// echoCommand - ECHO [<arg1> <arg2>]
func echoCommand(c Context) {
	c.WriteString(strings.Join(c.args, " "))
}
