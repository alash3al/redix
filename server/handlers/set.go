package handlers

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/alash3al/redix/server"
)

func init() {
	server.Handlers["set"] = server.Handler{
		Title:       "set",
		Description: "upsert a new key value pair",
		Examples: []string{
			"set key 'value'",
			"set key 'value' EX 12",
			"set key 'value' EX 12 NX",
			"set key 'value' EX 12 XX",
		},
		Group: "string",
		Callback: func(c *server.Context) error {
			args := c.Args()

			if len(args) < 2 {
				return errors.New("not enough argument specified")
			}

			key, value := append([]byte("data/strings/keys/"), args[0]...), args[1]

			opts := args[1:]
			ex, px := int64(0), int64(0)
			nx, xx := false, false
			exists, _ := c.DB().Has(key)

			for i, val := range opts {
				if bytes.EqualFold(val, []byte("ex")) || bytes.EqualFold(val, []byte("px")) {
					if (len(opts) - 1) < (i + 1) {
						return errors.New("you must specifiy the ttl")
					} else if bytes.EqualFold(val, []byte("ex")) {
						ex, _ = strconv.ParseInt(string(opts[i+1]), 10, 64)
					} else {
						px, _ = strconv.ParseInt(string(opts[i+1]), 10, 64)
					}
				}

				if bytes.EqualFold(val, []byte("nx")) {
					nx = true
				} else if bytes.EqualFold(val, []byte("xx")) {
					xx = true
				}
			}

			ttl := time.Duration(-1)
			if ex > 0 {
				ttl = time.Second * time.Duration(ex)
			} else if px > 0 {
				ttl = time.Second * time.Duration(px)
			}

			put := false

			if nx && !exists {
				put = true
			} else if xx && exists {
				put = true
			} else if !xx && !nx {
				put = true
			}

			expires := int64(-1)
			if int64(ttl) > 1 {
				expires = time.Now().Add(ttl).UnixNano()
			}

			value = append([]byte(fmt.Sprintf("%d;", expires)), value...)

			if put {
				if err := c.DB().Put(key, value); err != nil {
					return err
				}
			}

			c.Conn().WriteString("OK")

			return nil
		},
	}
}
