package handlers

import (
	"github.com/alash3al/redix/server"
)

func init() {
	server.Handlers["quit"] = server.Handler{
		Title:       "quit",
		Description: "closes the current connection",
		Examples: []string{
			"quit",
		},
		Group: "connection",
		Callback: func(c *server.Context) error {
			return c.Conn().Close()
		},
	}
}
