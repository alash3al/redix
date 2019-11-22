package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/alash3al/redix/db/driver"

	"github.com/alash3al/redix/server"
)

func init() {
	server.Handlers["del"] = server.Handler{
		Title:       "del",
		Description: "remove key(s) from the database",
		Examples:    []string{"del key1 key2 ..."},
		Group:       "genric",
		Callback: func(c *server.Context) error {
			args := c.Args()

			if len(args) < 1 {
				c.Conn().WriteInt(0)
				return nil
			}

			pairs := []driver.KeyValue{}

			for _, k := range args {
				k = append([]byte("data/strings/keys/"), k...)
				pairs = append(pairs, driver.KeyValue{Key: k, Value: nil})
			}

			if err := c.DB().Batch(pairs); err != nil {
				return err
			}

			c.Conn().WriteInt(len(pairs))
			return nil
		},
	}

	server.Handlers["exists"] = server.Handler{
		Title:       "exists",
		Description: "check whether specified key(s) already in the database",
		Examples:    []string{"exists key1 key2 ..."},
		Group:       "genric",
		Callback: func(c *server.Context) error {
			args := c.Args()

			if len(args) < 1 {
				c.Conn().WriteInt(0)
				return nil
			}

			i := 0

			for _, k := range args {
				k = append([]byte("data/strings/keys/"), k...)
				if ok, _ := c.DB().Has(k); ok {
					i++
				}
			}

			c.Conn().WriteInt(i)
			return nil
		},
	}

	server.Handlers["get"] = server.Handler{
		Title:       "get",
		Description: "fetches a key",
		Examples:    []string{"get foobar"},
		Group:       "string",
		Callback: func(c *server.Context) error {
			args := c.Args()
			if len(args) != 1 {
				return errors.New("ERR wrong number of arguments for 'get' command")
			}

			key := append([]byte("data/strings/keys/"), args[0]...)

			val, err := c.DB().Get(key)
			if err != nil {
				c.Conn().WriteNull()
				return nil
			}

			parts := bytes.SplitN(val, []byte(";"), 2)
			if len(parts) < 2 {
				c.Conn().WriteNull()
				return nil
			}

			val = parts[1]

			expires, _ := strconv.ParseInt(string(parts[0]), 10, 64)
			if expires > 0 && time.Now().UnixNano() >= expires {
				c.DB().Delete(key)
				c.Conn().WriteNull()
				return nil
			}

			c.Conn().WriteString(string(val))

			return nil
		},
	}

	server.Handlers["set"] = server.Handler{
		Title:       "set",
		Description: "upsert a new key value pair",
		Examples: []string{
			"set key 'value'",
			"set key 'value' EX 12",
			"set key 'value' EX 12 NX",
			"set key 'value' EX 12 XX",
		},
		Group: "string",
		Callback: func(c *server.Context) error {
			args := c.Args()

			if len(args) < 2 {
				return errors.New("not enough argument specified")
			}

			key, value := append([]byte("data/strings/keys/"), args[0]...), args[1]

			opts := args[1:]
			ex, px := int64(0), int64(0)
			nx, xx := false, false
			exists, _ := c.DB().Has(key)

			for i, val := range opts {
				if bytes.EqualFold(val, []byte("ex")) || bytes.EqualFold(val, []byte("px")) {
					if (len(opts) - 1) < (i + 1) {
						return errors.New("you must specifiy the ttl")
					} else if bytes.EqualFold(val, []byte("ex")) {
						ex, _ = strconv.ParseInt(string(opts[i+1]), 10, 64)
					} else {
						px, _ = strconv.ParseInt(string(opts[i+1]), 10, 64)
					}
				}

				if bytes.EqualFold(val, []byte("nx")) {
					nx = true
				} else if bytes.EqualFold(val, []byte("xx")) {
					xx = true
				}
			}

			ttl := time.Duration(-1)
			if ex > 0 {
				ttl = time.Second * time.Duration(ex)
			} else if px > 0 {
				ttl = time.Second * time.Duration(px)
			}

			put := false

			if nx && !exists {
				put = true
			} else if xx && exists {
				put = true
			} else if !xx && !nx {
				put = true
			}

			expires := int64(-1)
			if int64(ttl) > 1 {
				expires = time.Now().Add(ttl).UnixNano()
			}

			value = append([]byte(fmt.Sprintf("%d;", expires)), value...)

			if put {
				if err := c.DB().Put(key, value); err != nil {
					return err
				}
			}

			c.Conn().WriteString("OK")

			return nil
		},
	}
}
