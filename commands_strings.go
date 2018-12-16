package main

import (
	"fmt"
	"strconv"
)

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

// existsCommand - Exists <key>
func existsCommand(c Context) {
	if len(c.args) < 1 {
		c.WriteError("EXISTS command must has at least 1 argument EXISTS <key>")
		return
	}

	_, err := c.db.Get(c.args[0])
	if err != nil {
		c.WriteInt(0)
		return
	}

	c.WriteInt(1)
}

// incrCommand -
func incrCommand(c Context) {
	var key string
	var by int64

	if len(c.args) < 1 {
		c.WriteError("INCR command must has at least one argument INCR <key> [number]")
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
