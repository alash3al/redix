package commands

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/alash3al/redix/internals/datastore/contract"
)

func init() {
	// PING
	HandleFunc("ping", func(c *Context) {
		c.Conn.WriteString("PONG")
	})

	// QUIT
	HandleFunc("quit", func(c *Context) {
		c.Conn.WriteString("OK")
		c.Conn.Close()
	})

	// SELECT <DB index>
	HandleFunc("select", func(c *Context) {
		if c.Argc < 1 {
			c.Conn.WriteError("Err invalid arguments supplied")
			return
		}

		i, err := strconv.Atoi(string(c.Argv[0]))
		if err != nil {
			c.Conn.WriteError("Err invalid DB index")
			return
		}

		c.SessionSet("namespace", fmt.Sprintf("/%d/", i))

		c.Conn.WriteString("OK")
	})

	// GET <key> [DELETE]
	HandleFunc("get", func(c *Context) {
		if c.Argc < 1 {
			c.Conn.WriteError("Err invalid arguments specified")
			return
		}

		delete := false

		for i := 1; i < c.Argc; i++ {
			switch strings.ToLower(string(c.Argv[i])) {
			case "delete":
				delete = true
			}
		}

		ret, err := c.Engine.Read(&contract.ReadInput{
			Key:    c.AbsoluteKeyPath(c.Argv[0]),
			Delete: delete,
		})

		if err != nil {
			c.Conn.WriteError("Err " + err.Error())
			return
		}

		if len(ret.Value) < 1 {
			c.Conn.WriteNull()
			return
		}

		c.Conn.WriteBulk(ret.Value)
	})

	// GETDEL <key> =
	// same as: GET <key> DELETE
	HandleFunc("getdel", func(c *Context) {
		if c.Argc != 1 {
			c.Conn.WriteError("Err invalid number of arguments specified")
			return
		}

		c.Argc++
		c.Argv = append(c.Argv, []byte("DELETE"))

		Call("get", c)
	})

	// SET <key> <value> [EX seconds | KEEPTTL] [NX]
	HandleFunc("set", func(c *Context) {
		if c.Argc < 2 {
			c.Conn.WriteError("Err invalid arguments specified")
			return
		}

		writeOpts := contract.WriteInput{
			Key:   c.AbsoluteKeyPath(c.Argv[0]),
			Value: c.Argv[1],
		}

		if c.Argc > 2 {
			for i := 0; i < len(c.Argv); i++ {
				chr := string(bytes.ToLower(c.Argv[i]))
				switch chr {
				case "ex":
					n, err := strconv.ParseInt(string(c.Argv[i+1]), 10, 64)
					if err != nil {
						c.Conn.WriteError("Err " + err.Error())
						return
					}
					writeOpts.TTL = time.Second * time.Duration(n)
				case "keepttl":
					writeOpts.KeepTTL = true
				case "nx":
					writeOpts.OnlyIfNotExists = true
				}

			}
		}

		if c.Cfg.Server.Redis.AsyncWrites {
			go (func() {
				if _, err := c.Engine.Write(&writeOpts); err != nil {
					log.Println("[FATAL]", err.Error())
				}
			})()
		} else {
			if _, err := c.Engine.Write(&writeOpts); err != nil {
				c.Conn.WriteError("Err " + err.Error())
				return
			}
		}

		c.Conn.WriteString("OK")
	})

	// TTL <key>
	HandleFunc("ttl", func(c *Context) {
		if c.Argc < 1 {
			c.Conn.WriteError("Err invalid arguments specified")
			return
		}

		ret, err := c.Engine.Read(&contract.ReadInput{
			Key: c.AbsoluteKeyPath(c.Argv[0]),
		})

		if err != nil {
			c.Conn.WriteError("Err " + err.Error())
			return
		}

		if !ret.Exists {
			c.Conn.WriteBulkString("-2")
			return
		}

		if ret.TTL == 0 {
			c.Conn.WriteBulkString("-1")
			return
		}

		c.Conn.WriteAny(ret.TTL.Milliseconds())
	})

	// INCR <key> [<delta>]
	HandleFunc("incr", func(c *Context) {
		if c.Argc < 1 {
			c.Conn.WriteError("Err invalid arguments specified")
			return
		}

		delta := []byte("1")
		if c.Argc > 1 {
			delta = c.Argv[1]
		}

		if c.Cfg.Server.Redis.AsyncWrites {
			go (func() {
				if _, err := c.Engine.Write(&contract.WriteInput{
					Key:       c.AbsoluteKeyPath(c.Argv[0]),
					Value:     delta,
					Increment: true,
				}); err != nil {
					log.Println("[FATAL]", err.Error())
				}
			})()

			c.Conn.WriteNull()
			return
		}

		ret, err := c.Engine.Write(&contract.WriteInput{
			Key:       c.AbsoluteKeyPath(c.Argv[0]),
			Value:     delta,
			Increment: true,
		})

		if err != nil {
			c.Conn.WriteError("Err " + err.Error())
			return
		}

		c.Conn.WriteBulk(ret.Value)
	})

	// INCRBY <key> <delta>
	HandleFunc("incrby", func(c *Context) {
		Call("incr", c)
	})

	// DEL key [key ...]
	HandleFunc("del", func(c *Context) {
		if c.Argc < 1 {
			c.Conn.WriteError("Err invalid arguments specified")
			return
		}

		if c.Cfg.Server.Redis.AsyncWrites {
			go (func() {
				for i := range c.Argv {
					_, err := c.Engine.Write(&contract.WriteInput{
						Key:   c.AbsoluteKeyPath(c.Argv[i]),
						Value: nil,
					})

					if err != nil {
						log.Println("[FATAL]", err.Error())
						return
					}
				}
			})()

			c.Conn.WriteString("OK")
			return
		}

		for i := range c.Argv {
			_, err := c.Engine.Write(&contract.WriteInput{
				Key:   c.AbsoluteKeyPath(c.Argv[i]),
				Value: nil,
			})

			if err != nil {
				c.Conn.WriteError("Err " + err.Error())
				return
			}
		}

		c.Conn.WriteString("OK")
	})

	// HGETALL <prefix>
	HandleFunc("hgetall", func(c *Context) {
		prefix := []byte("")

		if c.Argc > 0 {
			prefix = c.Argv[0]
		}

		result := map[string]string{}

		err := c.Engine.Iterate(&contract.IteratorOpts{
			Prefix: c.AbsoluteKeyPath(prefix),
			Callback: func(ro *contract.ReadOutput) error {
				endKey := strings.TrimPrefix(string(ro.Key), string(c.AbsoluteKeyPath(prefix)))
				result[endKey] = string(ro.Value)
				return nil
			},
		})

		if err != nil && err != contract.ErrStopIterator {
			c.Conn.WriteError("ERR " + err.Error())
		}

		c.Conn.WriteAny(result)
	})

	// FLUSHALL
	HandleFunc("flushall", func(c *Context) {
		_, err := c.Engine.Write(&contract.WriteInput{
			Key:   nil,
			Value: nil,
		})

		if err != nil {
			c.Conn.WriteError("Err " + err.Error())
			return
		}

		c.Conn.WriteString("OK")
	})

	// FLUSHDB
	HandleFunc("flushdb", func(c *Context) {
		_, err := c.Engine.Write(&contract.WriteInput{
			Key:   c.AbsoluteKeyPath(),
			Value: nil,
		})

		if err != nil {
			c.Conn.WriteError("Err " + err.Error())
			return
		}

		c.Conn.WriteString("OK")
	})
}
