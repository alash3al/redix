package server

import (
	"fmt"
	"strings"

	"github.com/alash3al/redix/db"
	"github.com/tidwall/redcon"
)

// Options a server related options
type Options struct {
	RESPAddr string
	Openner  OpenFunc
}

// OpenFunc a database selector function
type OpenFunc func(dbname string) (*db.DB, error)

// ListenAndServe start listening and serving the incomming requests
func ListenAndServe(opts Options) error {
	return redcon.ListenAndServe(
		opts.RESPAddr,
		func(incommingConn redcon.Conn, incommingCommand redcon.Command) {
			defer (func() {
				if err := recover(); err != nil {
					incommingConn.WriteError(fmt.Sprintf("fatal error: %s", (err.(error)).Error()))
				}
			})()

			if len(incommingCommand.Args) < 1 {
				incommingConn.WriteError(errNoCommand.Error())
				return
			}

			commandName := strings.ToLower(string(incommingCommand.Args[0]))
			handler, ok := Handlers[commandName]

			if !ok {
				incommingConn.WriteError(errUnknownCommand.Error())
				return
			}

			ctx, ok := incommingConn.Context().(*Context)
			if !ok {
				incommingConn.WriteError("unexpected thing happened")
				return
			}

			ctx.args = incommingCommand.Args[1:]

			if err := handler.Callback(ctx); err != nil {
				incommingConn.WriteError(err.Error())
				return
			}
		},
		func(conn redcon.Conn) bool {
			defaultDB, err := opts.Openner("0")
			if err != nil {
				conn.WriteError(err.Error())
				return false
			}

			conn.SetContext(&Context{
				conn:       conn,
				serverOpts: opts,
				db:         defaultDB,
			})

			return true
		},
		nil,
	)
}
