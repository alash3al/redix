package commands

import (
	"strconv"
	"time"

	"github.com/alash3al/goukv"
	"github.com/alash3al/redix/internals/db"
	"github.com/alash3al/redix/internals/resp"
)

func init() {
	prefix := usrPrefix

	resp.Handlers["set"] = func(c *resp.Context) {
		args := c.Args()

		if len(args) < 2 {
			c.Conn().WriteError(incorrectArgsCount)
			return
		}

		entry := db.Entry{
			Key:   append(prefix, args[0]...),
			Value: args[1],
		}

		if len(args) > 2 {
			ttlSec, err := strconv.Atoi(string(args[2]))
			if err != nil {
				c.Conn().WriteError(err.Error())
				return
			}
			entry.TTL = time.Second * time.Duration(ttlSec)
		}

		c.DB().Put(&entry)
		c.Conn().WriteString("OK")
	}

	resp.Handlers["get"] = func(c *resp.Context) {
		args := c.Args()

		if len(args) < 1 {
			c.Conn().WriteError(incorrectArgsCount)
			return
		}

		k := append(prefix, args[0]...)
		val, err := c.DB().Get(k)

		if err == goukv.ErrKeyExpired {
			err = nil
			entry := db.Entry{Key: k}
			c.DB().Put(&entry)
		}

		if err != nil {
			c.Conn().WriteError(err.Error())
			return
		}

		if val == nil {
			c.Conn().WriteNull()
			return
		}

		c.Conn().WriteBulk((val))
	}

	resp.Handlers["del"] = func(c *resp.Context) {
		args := c.Args()

		if len(args) < 1 {
			c.Conn().WriteInt(0)
			return
		}

		entries := []*db.Entry{}
		for _, k := range args {
			entry := db.Entry{Key: append(prefix, k...)}
			entries = append(entries, &entry)
		}

		c.DB().Batch(entries)
		c.Conn().WriteInt(len(entries))
	}

	resp.Handlers["incr"] = func(c *resp.Context) {
		args := c.Args()

		if len(args) < 1 {
			c.Conn().WriteError(incorrectArgsCount)
			return
		}

		entry := db.Entry{
			Key: append(prefix, args[0]...),
		}

		oldValueBin, _ := c.DB().Get(entry.Key)
		oldValueNum, _ := strconv.ParseFloat(string(oldValueBin), 64)

		if len(args) > 1 {
			delta, err := strconv.ParseFloat(string(args[1]), 64)
			if err != nil {
				c.Conn().WriteError(err.Error())
				return
			}
			if delta == 0 {
				delta = 1
			}
			oldValueNum += float64(delta)
		} else {
			oldValueNum += 1
		}

		newValueString := strconv.FormatFloat(oldValueNum, 'f', -1, 64)
		entry.Value = []byte(newValueString)

		c.DB().Put(&entry)

		if float64(int64(oldValueNum)) == oldValueNum {
			c.Conn().WriteInt64(int64(oldValueNum))
		} else {
			c.Conn().WriteBulkString(newValueString)
		}
	}
}
