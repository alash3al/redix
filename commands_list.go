package main

import (
	"encoding/hex"
	"strconv"

	"github.com/rs/xid"
)

// lpushCommand - LPUSH <LIST> <val1> [<val2> ...]
func lpushCommand(c Context) {
	var k string
	var vals []string
	var done []string

	if len(c.args) < 2 {
		c.WriteError("LPUSH command requires at least two arguments LPUSH <LIST> <value> [<value> ...]")
		return
	}

	k, vals = c.args[0], c.args[1:]

	for _, v := range vals {
		offset := xid.New().String()
		key := k + "/{LIST}/" + offset
		if err := c.db.Set(key, v, -1); err != nil {
			done = append(done, "")
		}
		done = append(done, key)
	}

	c.WriteArray(len(done))
	for _, v := range done {
		c.WriteBulkString(hex.EncodeToString([]byte(v)))
	}
}

// lpushuCommand - LPUSHU <LIST> <val1> [<val2> ...]
func lpushuCommand(c Context) {
	var k string
	var vals []string
	var done []string

	if len(c.args) < 2 {
		c.WriteError("LPUSHU command requires at least two arguments command requires at least two arguments LPUSHU <LIST> <value> [<value> ...]")
		return
	}

	k, vals = c.args[0], c.args[1:]

	for _, v := range vals {
		offset := hex.EncodeToString([]byte(v))
		key := k + "/{LIST}/" + offset
		if err := c.db.Set(key, v, -1); err != nil {
			done = append(done, "")
		}
		done = append(done, key)
	}

	c.WriteArray(len(done))
	for _, v := range done {
		c.WriteBulkString(hex.EncodeToString([]byte(v)))
	}
}

// lrange - LGETALL <LIST> [<offset> <size>]
func lrangeCommand(c Context) {
	var key, offset string
	var limit int

	if len(c.args) < 1 {
		c.WriteError("LGETALL must has at least 1 argument")
		return
	}

	key = c.args[0]
	prefix := key + "/{LIST}/"

	if len(c.args) > 1 {
		offset = c.args[1]
	}

	if len(c.args) > 2 {
		limit, _ = strconv.Atoi(c.args[2])
	}

	if offset == "" {
		offset = prefix
	} else {
		of, err := hex.DecodeString(offset)
		if err != nil {
			c.WriteError("invalid offset specified")
			return
		}
		offset = string(of)
	}

	data := []string{}
	err := c.db.Scan(ScannerOptions{
		Offset:      offset,
		Prefix:      prefix,
		FetchValues: true,
		Handler: func(k, v string) bool {
			if limit > 0 && (len(data) == limit) {
				return false
			}
			data = append(data, k, v)
			return true
		},
	})

	if err != nil {
		c.WriteError(err.Error())
		return
	}

	if len(data) == 0 {
		c.WriteNull()
		return
	}

	lastKey := ""
	if len(data) > 2 {
		lastKey = data[len(data)-2]
	}

	lastKey = hex.EncodeToString([]byte(lastKey))

	c.WriteArray(2)

	if lastKey != "" {
		c.WriteBulkString(lastKey)
	} else {
		c.WriteNull()
	}

	c.WriteArray(len(data) / 2)
	for i, v := range data {
		if i%2 == 0 {
			continue
		}
		c.WriteBulkString(v)
	}
}

// lremCommand - LREM <LIST> <val> [<val> <val> ...]
func lremCommand(c Context) {
	if len(c.args) < 2 {
		c.WriteError("LREM command requires at least 1 arguments LREM <key> <val> [<val> <val> ...]")
		return
	}

	key, vals := c.args[0], c.args[1:]
	prefix := key + "/{LIST}/"
	valsMap := map[string]bool{}
	keys := []string{}

	for _, v := range vals {
		valsMap[v] = true
	}

	err := c.db.Scan(ScannerOptions{
		Offset:      prefix,
		Prefix:      prefix,
		FetchValues: true,
		Handler: func(k, v string) bool {
			if valsMap[v] {
				keys = append(keys, k)
			}
			return true
		},
	})

	if err != nil {
		c.WriteError(err.Error())
		return
	}

	if err := c.db.Del(keys); err != nil {
		c.WriteError(err.Error())
		return
	}

	c.WriteInt(len(keys))
}

// lcountCommand - LCOUNT <LIST>
func lcountCommand(c Context) {
	if len(c.args) < 1 {
		c.WriteError("LCOUNT command must has at least 1 argument LCOUNT <LIST>")
		return
	}

	size := int64(0)
	prefix := c.args[0] + "/{LIST}/"
	c.db.Scan(ScannerOptions{
		Offset:      prefix,
		Prefix:      prefix,
		FetchValues: false,
		Handler: func(k, v string) bool {
			size++
			return true
		},
	})

	c.WriteInt64(size)
}
