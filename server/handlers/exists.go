package handlers

import (
	"github.com/alash3al/redix/server"
)

func init() {
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
}
