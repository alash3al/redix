package redis

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/alash3al/redix/internals/binlog"
	"github.com/alash3al/redix/internals/datastore/contract"
	"github.com/alash3al/redix/internals/manager"
	"github.com/tidwall/redcon"
)

// ListenAndServe start a redis server
func ListenAndServe(addr string) error {
	mngr, err := manager.New(&manager.Options{
		DataDir:       "./redixdata",
		DefaultEngine: "boltdb",
	})

	if err != nil {
		return err
	}

	return redcon.ListenAndServe(addr,
		func(conn redcon.Conn, cmd redcon.Command) {
			switch strings.ToLower(string(cmd.Args[0])) {
			default:
				conn.WriteError("ERR unknown command '" + string(cmd.Args[0]) + "'")
			case "ping":
				conn.WriteString("PONG")
			case "quit":
				conn.WriteString("OK")
				conn.Close()
			case "binlog":
				if len(cmd.Args) < 1 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}

				var offset []byte

				if len(cmd.Args) > 1 {
					offset = cmd.Args[1]
				}

				limit := 1

				if len(cmd.Args) > 2 {
					limit, _ = strconv.Atoi(string(cmd.Args[2]))
				}

				fetchedLogs := []*binlog.LogEntry{}

				if err := mngr.BinLog().ForEach(offset, false, func(l *binlog.LogEntry) bool {
					if len(fetchedLogs) >= limit {
						return false
					}

					fetchedLogs = append(fetchedLogs, l)

					return true
				}); err != nil {
					conn.WriteError("ERR " + err.Error())
					return
				}

				conn.WriteArray(len(fetchedLogs))

				for _, e := range fetchedLogs {
					jbytes, err := json.Marshal(e)
					if err != nil {
						conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
						return
					}
					conn.WriteBulk(jbytes)
				}
			case "set":
				if len(cmd.Args) != 3 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				if _, err := mngr.Put(&contract.PutInput{
					Key:   cmd.Args[1],
					Value: cmd.Args[2],
				}); err != nil {
					conn.WriteError("ERR " + err.Error())
					return
				}
				conn.WriteString("OK")
			case "get":
				if len(cmd.Args) != 2 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}

				val, err := mngr.Get(&contract.GetInput{
					Key: cmd.Args[1],
				})

				if err != nil {
					conn.WriteError("ERR " + err.Error())
					return
				}

				if val.Value == nil {
					conn.WriteNull()
				} else {
					conn.WriteBulk(val.Value)
				}
			case "del":
				if len(cmd.Args) != 3 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}
				if _, err := mngr.Put(&contract.PutInput{
					Key:   cmd.Args[1],
					Value: nil,
				}); err != nil {
					conn.WriteError("ERR " + err.Error())
					return
				}
				conn.WriteString("OK")
			}
		},
		func(conn redcon.Conn) bool {
			// Use this function to accept or deny the connection.
			// log.Printf("accept: %s", conn.RemoteAddr())
			return true
		},
		func(conn redcon.Conn, err error) {
			// This is called when the connection has been closed
			// log.Printf("closed: %s, err: %v", conn.RemoteAddr(), err)
		},
	)
}
