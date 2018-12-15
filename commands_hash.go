package main

import (
	"fmt"
	"strconv"
	"strings"
)

// hsetCommand - HSET <HASHMAP> <KEY> <VALUE> <TTL>
func hsetCommand(c Context) {
	var ns, k, v string
	var ttl int

	if len(c.args) < 3 {
		c.WriteError("HSET command requires at least three arguments HSET <hashmap> <key> <value> [<TTL>]")
		return
	}

	ns, k, v = c.args[0], c.args[1], c.args[2]

	if len(c.args) > 3 {
		ttl, _ = strconv.Atoi(c.args[3])
	}

	if err := c.db.Set(ns+"/{HASH}/"+k, v, ttl); err != nil {
		c.WriteError(err.Error())
		return
	}

	c.WriteInt(1)
}

// hgetCommand - HGET <HASHMAP> <KEY>
func hgetCommand(c Context) {
	var ns, k string

	if len(c.args) < 2 {
		c.WriteError("HGET command requires at least two arguments HGET <hashmap> <key>")
		return
	}

	ns, k = c.args[0], c.args[1]

	c.args = []string{ns + "/{HASH}/" + k}

	getCommand(c)
}

// hdelCommand - HDEL <HASHMAP> <key1> [<key2> ...]
func hdelCommand(c Context) {
	var ns string

	if len(c.args) < 2 {
		c.WriteError("HGET command requires at least two arguments HGET <hashmap> <key>")
		return
	}

	ns = c.args[0]
	keys := c.args[1:]

	for i, k := range keys {
		keys[i] = ns + "/{HASH}/" + k
	}

	c.args = keys

	delCommand(c)
}

// hgetallCommand - HGETALL <HASHMAP>
func hgetallCommand(c Context) {
	if len(c.args) < 1 {
		c.WriteError("HGETALL command requires at least one argument HGETALL <HASHMAP>")
	}

	prefix := c.args[0] + "/{HASH}/"
	data := map[string]string{}
	err := c.db.Scan(ScannerOptions{
		FetchValues:   true,
		IncludeOffset: true,
		Prefix:        prefix,
		Offset:        prefix,
		Handler: func(k, v string) bool {
			p := strings.SplitN(k, "/{HASH}/", 2)
			if len(p) < 2 {
				return true
			}
			data[p[1]] = v
			return true
		},
	})

	if err != nil {
		c.WriteError(err.Error())
		return
	}

	c.WriteArray(len(data) * 2)

	for k, v := range data {
		c.WriteBulkString(k)
		c.WriteBulkString(v)
	}
}

// hmsetCommand - HMSET <HASHMAP> <key1> <val1> [<key2> <val2> ...]
func hmsetCommand(c Context) {
	var ns string

	if len(c.args) < 3 {
		c.WriteError("HMSET command requires at least three arguments HMSET <HASHMAP> <key1> <val1> [<key2> <val2> ...]")
		return
	}

	ns = c.args[0]
	args := c.args[1:]

	currentCount := len(args)
	if len(args)%2 != 0 {
		c.WriteError(fmt.Sprintf("HMSET {key => value} pairs must be even you specified %d, it should be %d or %d", currentCount, currentCount+1, currentCount-1))
		return
	}

	data := map[string]string{}
	for i, v := range args {
		index := i + 1
		if index%2 == 0 {
			data[ns+"/{HASH}/"+args[i-1]] = v
		} else {
			data[ns+"/{HASH}/"+args[i]] = ""
		}
	}

	if err := c.db.MSet(data); err != nil {
		c.WriteError(err.Error())
		return
	}

	c.WriteInt(len(data))
}

func hexistsCommand(c Context) {
	if len(c.args) < 1 {
		c.WriteError("HEXISTS command requires at least 1 argument HEXISTS <HASHMAP> [<key>]")
		return
	}

	ns := c.args[0]

	if len(c.args) > 1 {
		c.args = []string{ns + "/{HASH}/" + c.args[1]}
		existsCommands(c)
		return
	}

	found := 0
	prefix := ns + "/{HASH}/"

	c.db.Scan(ScannerOptions{
		Prefix: prefix,
		Offset: prefix,
		Handler: func(_, _ string) bool {
			found++
			return false
		},
	})

	c.WriteInt(found)
}
