package handlers

import (
	"bytes"
	"errors"
	"strconv"
	"time"

	"github.com/alash3al/redix/server"
)

func init() {
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
}
