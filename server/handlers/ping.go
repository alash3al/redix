package handlers

import (
	"bytes"
	"fmt"

	"github.com/alash3al/redix/server"
)

func init() {
	server.Handlers["ping"] = server.Handler{
		Title:       "ping",
		Description: "a ping <-> pong command",
		Examples:    []string{"ping foobar"},
		Group:       "connection",
		Callback: func(c *server.Context) error {
			c.Conn().WriteString(fmt.Sprintf("PONG %s", bytes.Join(c.Args(), []byte(" "))))
			return nil
		},
	}
}
