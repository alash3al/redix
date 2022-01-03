package redis

import (
	"strconv"
	"strings"

	"github.com/alash3al/redix/internals/datastore/contract"
	"github.com/alash3al/redix/internals/manager"
	"github.com/alash3al/redix/internals/wal"
	"github.com/tidwall/redcon"
)

// ListenAndServe start a redis server
func ListenAndServe(addr string, mngr *manager.Manager) error {
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
			case "setclusterwaloffset":
				if len(cmd.Args) != 2 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}

				if err := mngr.UpdateClusterMinimumOffset(string(cmd.Args[1])); err != nil {
					conn.WriteError("ERR updating the cluster minimum offset due to: " + err.Error())
					return
				}

				conn.WriteString("OK")
			case "waloffset":
				offset, err := mngr.CurrentOffset()
				if err != nil {
					conn.WriteError("ERR fetching the current data offset due to: " + err.Error())
					return
				}

				conn.WriteString(offset)
			case "wal":
				if len(cmd.Args) < 1 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}

				limit := 1

				if len(cmd.Args) > 1 {
					limit, _ = strconv.Atoi(string(cmd.Args[1]))
				}

				var offset []byte

				if len(cmd.Args) > 2 {
					offset = cmd.Args[2]
				}

				fetchedLogs := [][][]byte{}

				if err := mngr.Wal().Range(func(key, value []byte) bool {
					fetchedLogs = append(fetchedLogs, [][]byte{key, value})

					return true
				}, &wal.RangeOpts{
					Offset:             offset,
					IncludeOffsetValue: false,
					Limit:              int64(limit),
				}); err != nil {
					conn.WriteError("ERR " + err.Error())
					return
				}

				conn.WriteArray(len(fetchedLogs) * 2)

				for _, val := range fetchedLogs {
					conn.WriteBulk(val[0])
					conn.WriteBulk(val[1])
				}
			case "set":
				if len(cmd.Args) != 3 {
					conn.WriteError("ERR wrong number of arguments for '" + string(cmd.Args[0]) + "' command")
					return
				}

				if err := mngr.Write(&contract.WriteInput{
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
				if err := mngr.Write(&contract.WriteInput{
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
