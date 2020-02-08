package commands

import (
	"strconv"

	"github.com/alash3al/redix/internal/server/resp"
)

func init() {
	resp.Handlers["set"] = resp.Handler{
		Title:       "SET",
		Description: "assign a value to the specified key with optional ttl in seconds",
		Examples: []string{
			"set mykey myvalue",
			"set mykey myvalue 10",
		},
		Callback: func(c *resp.Context) {
			args := c.Args()

			if len(args) < 2 {
				c.Conn().WriteError("invalid number of argument supplied")
				return
			}

			k, v, ttl := args[0], args[1], 0
			if len(args) > 2 {
				ttl, _ = strconv.Atoi(string(args[2]))
			}

			if err := c.Container().Set(k, v, ttl); err != nil {
				c.Conn().WriteError(err.Error())
				return
			}

			c.Conn().WriteString("OK")
		},
	}
	resp.Handlers["get"] = resp.Handler{
		Title:       "GET",
		Description: "fetch the value of the specified key",
		Examples: []string{
			"get mykey",
		},
		Callback: func(c *resp.Context) {
			args := c.Args()

			if len(args) < 1 {
				c.Conn().WriteError("invalid number of argument supplied")
				return
			}

			if val, err := c.Container().Get(args[0]); err != nil {
				c.Conn().WriteError(err.Error())
			} else if val != nil {
				c.Conn().WriteString(string(val))
			} else {
				c.Conn().WriteNull()
			}
		},
	}
}
