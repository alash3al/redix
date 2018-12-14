package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rs/xid"
)

// pingCommand - PING
func pingCommand(c Context) {
	c.WriteString("PONG")
}

// setCommand - SET <key> <value> [<TTL "millisecond">]
func setCommand(c Context) {
	var k, v, ttl string
	if len(c.args) < 2 {
		c.WriteError("SET command requires at least two arguments SET <key> <value> [TTL Millisecond]")
		return
	}

	k, v = c.args[0], c.args[1]

	if len(c.args) > 2 {
		ttl = c.args[2]
	}

	ttlVal, _ := strconv.Atoi(ttl)

	if err := c.db.Set(k, v, ttlVal); err != nil {
		c.WriteError(err.Error())
		return
	}

	c.WriteString("OK")
}

// getCommand - GET <key> [<default value>]
func getCommand(c Context) {
	if len(c.args) < 1 {
		c.WriteError("GET command must has at least 1 arguments")
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
		c.WriteError("MGET command must has at least 1 argumentss")
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
		c.WriteError("DEL command must has at least 1 arguments")
		return
	}

	if err := c.db.Del(c.args); err != nil {
		c.WriteError(err.Error())
		return
	}

	c.WriteString("OK")
}

// scanCommand - SCAN [cursor] [keys] [limit]
func scanCommand(c Context) {
	var offset, keyOnly string
	var limit int

	if len(c.args) > 0 {
		offset = c.args[0]
	}

	if len(c.args) > 1 {
		keyOnly = strings.ToLower(c.args[1])
	}

	if len(c.args) > 2 {
		limit, _ = strconv.Atoi(c.args[2])
	}

	data, err := c.db.Scan(offset, keyOnly == "keys", limit)
	if err != nil {
		c.WriteError(err.Error())
		return
	}

	c.WriteArray(len(data))
	for _, v := range data {
		c.WriteBulkString(v)
	}
}

// msetCommand - MSET <key1> <value1> [<key2> <value2> ...]
func msetCommand(c Context) {
	currentCount := len(c.args)
	if len(c.args)%2 != 0 {
		c.WriteError(fmt.Sprintf("MSET command arguments must be even you specified %d, it should be %d or %d", currentCount, currentCount+1, currentCount-1))
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

// appendCommand - APPEND <key> <value> [<TTL>]
func appendCommand(c Context) {
	var k, v, ttl string
	if len(c.args) < 2 {
		c.WriteError("APPEND command requires at least two arguments SET <key> <value> <TTL Millisecond>")
		return
	}

	k, v = c.args[0], c.args[1]

	if len(c.args) > 2 {
		ttl = c.args[2]
	}

	ttlVal, _ := strconv.Atoi(ttl)
	offset := xid.New().String()

	if err := c.db.Set(k+"/{ARRAY}/"+offset, v, ttlVal); err != nil {
		c.WriteError(err.Error())
		return
	}

	c.WriteInt(1)
}

// mappendCommand - MAPPEND <key> <val1> [<val2> ...]
func mappendCommand(c Context) {
	var k string
	var vals []string
	var done int

	if len(c.args) < 2 {
		c.WriteError("MAPPEND command requires at least two arguments SET <key> <value> <TTL Millisecond>")
		return
	}

	k, vals, done = c.args[0], c.args[1:], 0

	for _, v := range vals {
		offset := xid.New().String()
		if err := c.db.Set(k+"/{ARRAY}/"+offset, v, -1); err != nil {
			continue
		}
		done++
	}

	c.WriteInt(done)
}

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

	offset := c.args[0] + "/{HASH}/%"
	data, err := c.db.Scan(offset, false, -1)
	if err != nil {
		c.WriteError(err.Error())
		return
	}

	c.WriteArray(len(data))
	i := -1
	for _, v := range data {
		i++
		if i%2 == 0 {
			p := strings.SplitN(v, "/{HASH}/", 2)
			if len(p) < 2 {
				p = append(p, "")
			}
			v = p[1]
		}
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
