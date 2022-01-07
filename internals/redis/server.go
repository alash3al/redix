package redis

import (
	"fmt"
	"log"
	"sync/atomic"

	"github.com/alash3al/redix/internals/config"
	"github.com/alash3al/redix/internals/datastore/contract"
	"github.com/alash3al/redix/internals/redis/commands"
	"github.com/tidwall/redcon"
)

var (
	connCounter int64 = 0
)

// ListenAndServe start a redis server
func ListenAndServe(cfg *config.Config, engine contract.Engine) error {
	commands.HandleFunc("CLIENTCOUNT", func(c *commands.Context) {
		c.Conn.WriteAny(atomic.LoadInt64(&connCounter))
	})

	fmt.Println("=> started listening on", cfg.Server.Redis.ListenAddr, "...")
	return redcon.ListenAndServe(cfg.Server.Redis.ListenAddr,
		func(conn redcon.Conn, cmd redcon.Command) {
			ctx := commands.Context{
				Conn:   conn,
				Engine: engine,
				Cfg:    cfg,
				Argc:   len(cmd.Args) - 1,
				Argv:   cmd.Args[1:],
			}

			commands.Call(string(cmd.Args[0]), &ctx)
		},
		func(conn redcon.Conn) bool {
			if cfg.Server.Redis.MaxConns > 0 && cfg.Server.Redis.MaxConns <= atomic.LoadInt64(&connCounter) {
				log.Println("max connections reached!")
				return false
			}

			atomic.AddInt64(&connCounter, 1)

			conn.SetContext(map[string]interface{}{
				"namespace": "/0/",
			})
			return true
		},
		func(conn redcon.Conn, err error) {
			atomic.AddInt64(&connCounter, -1)
		},
	)
}
