// Package redis represents redis server
package redis

import (
	"fmt"
	"strings"

	"github.com/alash3al/redix/configparser"
	"github.com/alash3al/redix/redis/commands"
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

		if !ctx.IsAuthenticated && commandName != "auth" {
			clientConn.WriteError("AUTH is required")
			return
		}

		if ctx.CurrentDatabase == 0 && ctx.IsAuthenticated && commandName != "select" {
			if _, err := (commands.Commands["select"])(ctx); err != nil {
				clientConn.WriteError(err.Error())
				return
			}
		}

		commandHandler, found := commands.Commands[commandName]
		if !found {
			clientConn.WriteError("command not found")
			return
		}

		result, err := commandHandler(ctx)
		if err != nil {
			clientConn.WriteError(err.Error())
			return
		}

		clientConn.SetContext(ctx)

		clientConn.WriteAny(result)
	}
}

func accept(clientConn redcon.Conn) bool {
	clientConn.SetContext(&ctx.Ctx{})
	return true
}
