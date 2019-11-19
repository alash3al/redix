package handlers

import (
	"github.com/alash3al/redix/server"
)

func init() {
	server.Handlers["dbname"] = server.Handler{
		Title:       "dbname",
		Description: "returns the current database name/index",
		Examples: []string{
			"dbname",
		},
		Group: "server",
		Callback: func(c *server.Context) error {
			c.Conn().WriteString(c.DB().Name())
			return nil
		},
	}
}
