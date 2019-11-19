package handlers

import (
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
}
