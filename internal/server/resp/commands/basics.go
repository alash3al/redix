package commands

import (
	"github.com/alash3al/redix/internal/server/resp"
)

func init() {
	resp.Handlers["set"] = resp.Handler{
		Callback: func(c *resp.Context) {
			args := c.Args()
			k, v := args[0], args[1]

			if err := c.Container().Set(k, v, 0); err != nil {
				c.Conn().WriteError(err.Error())
				return
			}


			c.Conn().WriteString("OK")
		},
	}
}
