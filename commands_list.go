package main

import (
	"encoding/hex"
	"regexp"
	"strconv"
	"strings"

	"github.com/alash3al/redix/kvstore"
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
	err := c.db.Scan(kvstore.ScannerOptions{
		IncludeOffset: true,
		Offset:        offset,
		Prefix:        prefix,
		FetchValues:   true,
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
	if len(c.args) < 1 {
		c.WriteError("LREM command requires at least 1 arguments LREM <key> [<val1> <val2> <val3> ...]")
		return
	}

	key, vals := c.args[0], c.args[1:]
	prefix := key + "/{LIST}/"
	valsMap := map[string]bool{}
	keys := []string{}

	for _, v := range vals {
		valsMap[v] = true
	}

	err := c.db.Scan(kvstore.ScannerOptions{
		Offset:        prefix,
		IncludeOffset: true,
		Prefix:        prefix,
		FetchValues:   true,
		Handler: func(k, v string) bool {
			if len(valsMap) < 1 || valsMap[v] {
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
	c.db.Scan(kvstore.ScannerOptions{
		Offset:        prefix,
		IncludeOffset: true,
		Prefix:        prefix,
		FetchValues:   false,
		Handler: func(k, v string) bool {
			size++
			return true
		},
	})

	c.WriteInt64(size)
}

// lsumCommand - LSUM <list>
func lsumCommand(c Context) {
	if len(c.args) < 1 {
		c.WriteError("LSUM command must has at least 1 argument LSUM <LIST>")
		return
	}

	sum := int64(0)
	prefix := c.args[0] + "/{LIST}/"
	c.db.Scan(kvstore.ScannerOptions{
		Offset:        prefix,
		IncludeOffset: true,
		Prefix:        prefix,
		FetchValues:   true,
		Handler: func(_, v string) bool {
			i, _ := strconv.ParseInt(v, 10, 64)
			sum += i
			return true
		},
	})

	c.WriteInt64(sum)
}

// lavgCommand - LSUM <list>
func lavgCommand(c Context) {
	if len(c.args) < 1 {
		c.WriteError("LSUM command must has at least 1 argument LSUM <LIST>")
		return
	}

	sum := int64(0)
	size := int64(0)
	prefix := c.args[0] + "/{LIST}/"
	c.db.Scan(kvstore.ScannerOptions{
		Offset:        prefix,
		IncludeOffset: true,
		Prefix:        prefix,
		FetchValues:   true,
		Handler: func(_, v string) bool {
			i, _ := strconv.ParseInt(v, 10, 64)
			sum += i
			size++
			return true
		},
	})

	c.WriteInt64(sum / size)
}

// lminCommand - LMIN <list>
func lminCommand(c Context) {
	if len(c.args) < 1 {
		c.WriteError("LMIN command must has at least 1 argument LMIN <LIST>")
		return
	}

	min := int64(0)
	started := false
	prefix := c.args[0] + "/{LIST}/"
	c.db.Scan(kvstore.ScannerOptions{
		Offset:        prefix,
		IncludeOffset: true,
		Prefix:        prefix,
		FetchValues:   true,
		Handler: func(_, v string) bool {
			i, _ := strconv.ParseInt(v, 10, 64)
			if !started {
				min = i
				started = true
			} else if i < min {
				min = i
			}
			return true
		},
	})

	c.WriteInt64(min)
}

// lmaxCommand - LMAX <list>
func lmaxCommand(c Context) {
	if len(c.args) < 1 {
		c.WriteError("LMAX command must has at least 1 argument LMAX <LIST>")
		return
	}

	max := int64(0)
	started := false
	prefix := c.args[0] + "/{LIST}/"
	c.db.Scan(kvstore.ScannerOptions{
		Offset:        prefix,
		IncludeOffset: true,
		Prefix:        prefix,
		FetchValues:   true,
		Handler: func(_, v string) bool {
			i, _ := strconv.ParseInt(v, 10, 64)
			if !started {
				max = i
				started = true
			} else if i > max {
				max = i
			}
			return true
		},
	})

	c.WriteInt64(max)
}

// lsearchCommand - LSRCH <list> <pattern>
func lsearchCommand(c Context) {
	if len(c.args) < 2 {
		c.WriteError("LSRCH command must has at least 2 argument LSRCH <LIST> <regexp>")
		return
	}

	re, err := regexp.CompilePOSIX(c.args[1])
	if err != nil {
		c.WriteError(err.Error())
		return
	}

	result := []string{}

	prefix := c.args[0] + "/{LIST}/"
	c.db.Scan(kvstore.ScannerOptions{
		Offset:        prefix,
		IncludeOffset: true,
		Prefix:        prefix,
		FetchValues:   true,
		Handler: func(_, v string) bool {
			v = strings.ToLower(v)
			if re.MatchString(v) || strings.Contains(v, strings.ToLower(c.args[1])) {
				result = append(result, v)
			}
			return true
		},
	})

	c.WriteArray(len(result))
	for _, v := range result {
		c.WriteBulkString(v)
	}
}

// lsearchcountCommand - LSRCHCOUNT <list> <pattern>
func lsearchcountCommand(c Context) {
	if len(c.args) < 2 {
		c.WriteError("LSRCHCOUNT command must has at least 2 argument LSRCHCOUNT <LIST> <regexp>")
		return
	}

	re, err := regexp.CompilePOSIX(c.args[1])
	if err != nil {
		c.WriteError(err.Error())
		return
	}

	result := int64(0)

	prefix := c.args[0] + "/{LIST}/"
	c.db.Scan(kvstore.ScannerOptions{
		Offset:        prefix,
		IncludeOffset: true,
		Prefix:        prefix,
		FetchValues:   true,
		Handler: func(_, v string) bool {
			v = strings.ToLower(v)
			if re.MatchString(v) || strings.Contains(v, strings.ToLower(c.args[1])) {
				result++
			}
			return true
		},
	})

	c.WriteInt64(result)
}
