// Package redis represents redis server
package redis

import (
	"bytes"
	"fmt"

	"github.com/alash3al/redix/configparser"
	"github.com/alash3al/redix/redis/commands"
	"github.com/alash3al/redix/redis/context"
	"github.com/alash3al/redix/redis/store"
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
		handler(engineConn),
		accept,
		nil,
	)
}

func handler(engineConn store.Store) func(clientConn redcon.Conn, cmd redcon.Command) {
	return func(clientConn redcon.Conn, cmd redcon.Command) {
		if len(cmd.Args) < 1 {
			clientConn.WriteError("no command specified")
			return
		}

		ctx := clientConn.Context().(*context.Context)
		ctx.Command = bytes.ToLower(cmd.Args[0])
		ctx.Args = cmd.Args[1:]

		if engineConn.IsAuthRequired() && !ctx.IsAuthorized && string(ctx.Command) != "auth" {
			clientConn.WriteError("AUTH is required")
			return
		}

		if engineConn.IsAuthRequired() && !ctx.IsAuthorized && string(ctx.Command) == "auth" {
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

			ctx.SetContext(ctx)
			clientConn.WriteString("OK")

			return
		}

		if err := engineConn.Select(ctx, "0"); err != nil {
			clientConn.WriteError(err.Error())
			return
		}

		ctx.SetContext(ctx)

		command, exists := commands.Commands[string(ctx.Command)]
		if !exists {
			clientConn.WriteError("NOT IMPLEMENTED")
			return
		}

		result, err := command(commands.Request{
			Context: ctx,
			Store:   engineConn,
		})
		if err != nil {
			clientConn.WriteError(err.Error())
			return
		}

		ctx.SetContext(ctx)

		clientConn.WriteAny(result)
	}
}

func accept(clientConn redcon.Conn) bool {
	ctx := context.Context{
		Conn: clientConn,
	}

	clientConn.SetContext(&ctx)

	return true
}
