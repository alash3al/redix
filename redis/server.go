// Package redis represents redis server
package redis

import (
	"bytes"
	"fmt"

	"github.com/alash3al/redix/configparser"
	"github.com/alash3al/redix/redis/context"
	"github.com/alash3al/redix/redis/store/engines"
	"github.com/tidwall/redcon"

	_ "github.com/alash3al/redix/redis/store/engines/postgres"
)

// ListenAndServe start a RESP listener and serve the incoming requests
func ListenAndServe(config *configparser.Config) error {
	fmt.Println("=> initializing redis storage engine:", config.Storage.Driver)
	engineConn, err := engines.OpenStorageEngine(config)
	if err != nil {
		return err
	}

	fmt.Println("=> starting redis server on:", config.Server.Redis.ListenAddr)

	return redcon.ListenAndServe(
		config.Server.Redis.ListenAddr,
		func(clientConn redcon.Conn, cmd redcon.Command) {
			if len(cmd.Args) < 1 {
				clientConn.WriteError("no command specified")
				return
			}

			ctx := context.Context{
				Conn:    clientConn,
				Command: bytes.ToLower(cmd.Args[0]),
				Args:    cmd.Args[1:],
			}

			if engineConn.AuthRequired() && !ctx.IsAuthorized && string(ctx.Command) != "auth" {
				clientConn.WriteError("AUTH is required")
				return
			}

			if engineConn.AuthRequired() && !ctx.IsAuthorized && string(ctx.Command) == "auth" {
				token := ""
				if len(ctx.Args) > 0 {
					token = string(ctx.Args[0])
				}

				if exists, err := engineConn.AuthValidate(token); err != nil {
					clientConn.WriteError(err.Error())
					return
				} else if !exists {
					clientConn.WriteError("AUTH mismatch")
					return
				}

				ctx.CurrentToken = token
				ctx.IsAuthorized = true

				clientConn.WriteString("OK")
				return
			}

			result, err := engineConn.Exec(&ctx)
			if err != nil {
				clientConn.WriteError(err.Error())
				return
			}

			clientConn.WriteAny(result)
		},
		nil, nil,
	)
}
