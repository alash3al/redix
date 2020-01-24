package handlers

import (
	"errors"
	"strconv"
	"sync/atomic"

	"github.com/alash3al/redix/pkg/db/driver"
	"github.com/alash3al/redix/server"
)

var (
	writes atomic.Value

	prefix = []byte("")
)

func init() {
	server.Handlers["get"] = server.Handler{
		Title:       "get",
		Description: "fetches a key",
		Examples:    []string{"get foobar"},
		Group:       "core",
		Callback: func(c *server.Context) error {
			args := c.Args()
			if len(args) != 1 {
				return errors.New("ERR wrong number of arguments for 'get' command")
			}

			key := args[0]
			pair, err := c.DB().Get(key)
			if err != nil {
				return err
			}

			c.Conn().WriteString(string(pair.Value))

			return nil
		},
	}

	server.Handlers["set"] = server.Handler{
		Title:       "set",
		Description: "upsert a new key value pair (and optionally a ttl in milliseconds)",
		Examples: []string{
			"set key 'value'",
			"set key 'value' 1000",
		},
		Group: "core",
		Callback: func(c *server.Context) error {
			args := c.Args()

			if len(args) < 2 {
				return errors.New("not enough argument specified")
			}

			key, value := args[0], args[1]
			ttl := -1

			if len(args) > 2 {
				ttl, _ = strconv.Atoi(string(args[2]))
			}

			c.DB().Put(driver.Pair{
				Key:   key,
				Value: value,
				TTL:   ttl,
				Async: true,
			})

			c.Conn().WriteString("OK")

			return nil
		},
	}

	server.Handlers["exists"] = server.Handler{
		Title:       "exists",
		Description: "check whether specified key(s) already in the database",
		Examples:    []string{"exists key1 key2 ..."},
		Group:       "core",
		Callback: func(c *server.Context) error {
			args := c.Args()

			if len(args) < 1 {
				c.Conn().WriteInt(0)
				return nil
			}

			i := 0

			for _, k := range args {
				k = append(prefix, k...)
				if ok, _ := c.DB().Has(k); ok {
					i++
				}
			}

			c.Conn().WriteInt(i)
			return nil
		},
	}

	server.Handlers["del"] = server.Handler{
		Title:       "del",
		Description: "remove key(s) from the database",
		Examples:    []string{"del key1 key2 ..."},
		Group:       "core",
		Callback: func(c *server.Context) error {
			args := c.Args()

			if len(args) < 1 {
				c.Conn().WriteInt(0)
				return nil
			}

			for _, k := range args {
				c.DB().Put(driver.Pair{
					Key:   append(prefix, k...),
					Async: true,
				})
			}

			c.Conn().WriteInt(len(args))

			return nil
		},
	}
}
