package handlers

import (
	"bytes"
	"fmt"

	"github.com/alash3al/redix/server"
)

func init() {
	server.Handlers["ping"] = server.Handler{
		Title:       "ping",
		Description: "just a noop command",
		Func: func(c server.Context) error {
			c.Conn().WriteString(fmt.Sprintf("PONG %s", bytes.Join(c.Command().Args[1:], []byte(" "))))
			return nil
		},
	}
}
