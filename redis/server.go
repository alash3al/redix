// Package redis represents redis server
package redis

import (
	"fmt"
	"strings"

	"github.com/alash3al/redix/configparser"
	"github.com/alash3al/redix/redis/ctx"
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

	// fmt.Println(engineConn.AuthCreate())

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

		commandName := strings.ToLower(string(cmd.Args[0]))
		args := []string{}

		for _, v := range cmd.Args[1:] {
			args = append(args, string(v))
		}

		ctx := clientConn.Context().(*ctx.Ctx)
		ctx.Conn = clientConn
		ctx.Store = engineConn
		ctx.Args = args
		ctx.Command = commandName
		ctx.CurrentDatabase = -1

		if !ctx.IsAuthenticated && commandName != "auth" {
			clientConn.WriteError("AUTH is required")
			return
		}

		if commandName == "auth" {
			if len(args) < 1 {
				clientConn.WriteError("auth token isn't provided")
				return
			}

			ok, err := engineConn.AuthValidate(args[0])
			if err != nil {
				clientConn.WriteError(err.Error())
				return
			}

			if !ok {
				clientConn.WriteError("invalid auth data")
				return
			}

			ctx.IsAuthenticated = true
			ctx.CurrentToken = args[0]

			clientConn.SetContext(ctx)

			clientConn.WriteAny("Ok")

			return
		}

		if ctx.CurrentDatabase < 0 && ctx.IsAuthenticated && commandName != "select" {
			if actualDB, err := engineConn.Select(ctx.CurrentToken, 0); err != nil {
				clientConn.WriteError(err.Error())
				return
			} else {
				ctx.CurrentDatabase = actualDB
			}
		}

		if commandName == "select" {
			if len(args) < 1 {
				clientConn.WriteError("database not specified")
				return
			}

			if actualDB, err := engineConn.Select(ctx.CurrentToken, 0); err != nil {
				clientConn.WriteError(err.Error())
				return
			} else {
				ctx.CurrentDatabase = actualDB
			}
		}

		clientConn.SetContext(ctx)

		result, err := engineConn.Exec(commandName, args...)
		if err != nil {
			clientConn.WriteError(err.Error())
			return
		}

		clientConn.WriteAny(result)
	}
}

func accept(clientConn redcon.Conn) bool {
	clientConn.SetContext(&ctx.Ctx{})
	return true
}
