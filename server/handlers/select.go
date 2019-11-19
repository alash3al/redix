package handlers

import (
	"errors"

	"github.com/alash3al/redix/server"
)

func init() {
	server.Handlers["select"] = server.Handler{
		Title:       "select",
		Description: "change the current database",
		Examples: []string{
			"select 1",
		},
		Group: "connection",
		Callback: func(c *server.Context) error {
			args := c.Args()

			if len(args) != 1 {
				return errors.New("this command accepts '1' argument")
			}

			_, err := c.ChangeDB(string(args[0]))
			if err != nil {
				return err
			}

			c.Conn().WriteString("OK")

			return nil
		},
	}
}
