// Copyright 2018 The Redix Authors. All rights reserved.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.
package main

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/alash3al/redix/kvstore"
)

// setCommand - SET <key> <value> [<TTL "millisecond">]
func setCommand(c Context) {
	var k, v, ttl string
	if len(c.args) < 2 {
		c.WriteError("SET command requires at least two arguments: SET <key> <value> [TTL Millisecond]")
		return
	}

	k, v = c.args[0], c.args[1]

	if len(c.args) > 2 {
		ttl = c.args[2]
	}

	ttlVal, _ := strconv.Atoi(ttl)
	if ttlVal < 0 {
		ttlVal = 0
	}

	// if *flagACK {
	if err := c.db.Set(k, v, ttlVal); err != nil {
		c.WriteError(err.Error())
		return
	}
	// } else {
	// 	kvjobs <- func() {
	// 		c.db.Set(k, v, ttlVal)
	// 	}
	// }

	c.WriteString("OK")
}

// getCommand - GET <key> [<default value>]
func getCommand(c Context) {
	if len(c.args) < 1 {
		c.WriteError("GET command must have at least 1 argument: GET <key> [default value]")
		return
	}

	defaultVal := ""
	data, err := c.db.Get(c.args[0])

	if len(c.args) > 1 {
		defaultVal = c.args[1]
	}

	if err != nil {
		if defaultVal != "" {
			c.WriteString(defaultVal)
		} else {
			c.WriteNull()
		}
		return
	}

	c.WriteString(string(data))
}

// mgetCommand - MGET <key1> [<key2> ...]
func mgetCommand(c Context) {
	if len(c.args) < 1 {
		c.WriteError("MGET command must have at least 1 argument: MGET <key1> [<key2> ...]")
		return
	}

	data := c.db.MGet(c.args)

	c.WriteArray(len(data))
	for _, v := range data {
		if v == "" {
			c.WriteNull()
			continue
		}
		c.WriteBulkString(v)
	}
}

// delCommand - DEL <key1> [<key2> ...]
func delCommand(c Context) {
	if len(c.args) < 1 {
		c.WriteError("DEL command must have at least 1 argument: DEL <key1> [<key2> ...]")
		return
	}

	if err := c.db.Del(c.args); err != nil {
		c.WriteError(err.Error())
		return
	}

	c.WriteString("OK")
}

// msetCommand - MSET <key1> <value1> [<key2> <value2> ...]
func msetCommand(c Context) {
	currentCount := len(c.args)
	if len(c.args)%2 != 0 {
		c.WriteError(fmt.Sprintf("MSET command arguments must be even. You specified %d, it should be %d or %d", currentCount, currentCount+1, currentCount-1))
		return
	}

	data := map[string]string{}

	for i, v := range c.args {
		index := i + 1
		if index%2 == 0 {
			data[c.args[i-1]] = v
		} else {
			data[c.args[i]] = ""
		}
	}

	if err := c.db.MSet(data); err != nil {
		c.WriteError(err.Error())
		return
	}

	c.WriteInt(len(data))
}

// existsCommand - Exists <key>
func existsCommand(c Context) {
	if len(c.args) < 1 {
		c.WriteError("EXISTS command must have at least 1 argument: EXISTS <key>")
		return
	}

	_, err := c.db.Get(c.args[0])
	if err != nil {
		c.WriteInt(0)
		return
	}

	c.WriteInt(1)
}

// incrCommand - INCR <key> [number]
func incrCommand(c Context) {
	var key string
	var by int64

	if len(c.args) < 1 {
		c.WriteError("INCR command must have at least one argument: INCR <key> [number]")
		return
	}

	key = c.args[0]

	if len(c.args) > 1 {
		by, _ = strconv.ParseInt(c.args[1], 10, 64)
	}

	if by == 0 {
		by = 1
	}

	val, err := c.db.Incr(key, by)
	if err != nil {
		c.WriteError(err.Error())
		return
	}

	c.WriteInt64(int64(val))
}

// ttlCommand - TTL <key>
func ttlCommand(c Context) {
	if len(c.args) < 1 {
		c.WriteError("TTL command requires at least 1 argument, TTL <key>")
		return
	}
	c.WriteInt64(int64(c.db.TTL(c.args[0])))
}

// keysCommand - KEYS [<regexp-pattern>]
func keysCommand(c Context) {
	var data []string
	var pattern *regexp.Regexp
	var err error

	if len(c.args) > 0 {
		pattern, err = regexp.CompilePOSIX(c.args[0])
	}

	if err != nil {
		c.WriteError(err.Error())
		return
	}

	err = c.db.Scan(kvstore.ScannerOptions{
		FetchValues:   false,
		IncludeOffset: true,
		Handler: func(k, _ string) bool {
			if pattern != nil && pattern.MatchString(k) {
				data = append(data, k)
			} else if nil == pattern {
				data = append(data, k)
			}
			return true
		},
	})

	if err != nil {
		c.WriteError(err.Error())
		return
	}

	c.WriteArray(len(data))
	for _, k := range data {
		c.WriteBulkString(k)
	}
}
