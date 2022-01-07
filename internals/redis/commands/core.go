package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alash3al/redix/internals/datastore/contract"
)

func init() {
	HandleFunc("ping", func(c *Context) {
		c.Conn.WriteString("PONG")
	})

	HandleFunc("quit", func(c *Context) {
		c.Conn.WriteString("OK")
		c.Conn.Close()
	})

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

	HandleFunc("incr", func(c *Context) {
		if c.Argc < 1 {
			c.Conn.WriteError("Err invalid arguments specified")
			return
		}

		ret, err := c.Engine.Write(&contract.WriteInput{
			Key:       c.AbsoluteKeyPath(c.Argv[0]),
			Value:     []byte("1"),
			Increment: true,
		})

		if err != nil {
			c.Conn.WriteError("Err " + err.Error())
			return
		}

		c.Conn.WriteBulk(ret.Value)
	})

	HandleFunc("get", func(c *Context) {
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

		if len(ret.Value) < 1 {
			c.Conn.WriteNull()
			return
		}

		c.Conn.WriteBulk(ret.Value)
	})

	HandleFunc("set", func(c *Context) {
		if c.Argc < 2 {
			c.Conn.WriteError("Err invalid arguments specified")
			return
		}

		_, err := c.Engine.Write(&contract.WriteInput{
			Key:   c.AbsoluteKeyPath(c.Argv[0]),
			Value: c.Argv[1],
		})

		if err != nil {
			c.Conn.WriteError("Err " + err.Error())
			return
		}

		c.Conn.WriteString("OK")
	})

	HandleFunc("del", func(c *Context) {
		if c.Argc < 1 {
			c.Conn.WriteError("Err invalid arguments specified")
			return
		}

		_, err := c.Engine.Write(&contract.WriteInput{
			Key:   c.AbsoluteKeyPath(c.Argv[0]),
			Value: nil,
		})

		if err != nil {
			c.Conn.WriteError("Err " + err.Error())
			return
		}

		c.Conn.WriteString("OK")
	})

	HandleFunc("hgetall", func(c *Context) {
		prefix := []byte("")

		if c.Argc > 0 {
			prefix = c.Argv[0]
		}

		result := map[string]string{}

		err := c.Engine.Iterate(&contract.IteratorOpts{
			Prefix: c.AbsoluteKeyPath(prefix),
			Callback: func(ro *contract.ReadOutput) error {
				endKey := strings.TrimPrefix(string(ro.Key), string(c.AbsoluteKeyPath()))
				result[endKey] = string(ro.Value)
				return nil
			},
		})

		if err != nil && err != contract.ErrStopIterator {
			c.Conn.WriteError("ERR " + err.Error())
		}

		c.Conn.WriteAny(result)
	})

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
