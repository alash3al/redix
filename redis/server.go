// Package redis represents redis server
package redis

import (
	"fmt"

	"github.com/alash3al/redix/configparser"
	"github.com/alash3al/redix/redis/context"
	"github.com/alash3al/redix/redis/store/engines"
	"github.com/tidwall/redcon"

	_ "github.com/alash3al/redix/redis/store/engines/postgres"
)

func ListenAndServe(config *configparser.Config) error {
	fmt.Println("=> initializing the storage engine: ", config.Engine)
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
				Command: cmd.Args[0],
				Args:    cmd.Args[1:],
			}

			clientConn.WriteAny(engineConn.Exec(ctx))
		},
		nil, nil,
	)
}
