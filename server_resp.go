// Copyright 2018 The Redix Authors. All rights reserved.
// Use of this source code is governed by a Apache 2.0
// license that can be found in the LICENSE file.
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/alash3al/go-color"
	"github.com/tidwall/redcon"
)

func initRespServer() error {
	return redcon.ListenAndServe(
		*flagRESPListenAddr,
		func(conn redcon.Conn, cmd redcon.Command) {
			// handles any panic
			defer (func() {
				if err := recover(); err != nil {
					conn.WriteError(fmt.Sprintf("fatal error: %s", (err.(error)).Error()))
				}
			})()

			// fetch the connection context
			// normalize the todo action "command"
			// normalize the command arguments
			ctx := (conn.Context()).(map[string]interface{})
			todo := strings.TrimSpace(strings.ToLower(string(cmd.Args[0])))
			args := []string{}
			for _, v := range cmd.Args[1:] {
				v := strings.TrimSpace(string(v))
				args = append(args, v)
			}

			// verbose ?
			if *flagVerbose {
				log.Println(color.YellowString(todo), color.CyanString(strings.Join(args, " ")))
			}

			// internal command to pick a database
			if todo == "select" {
				if len(args) < 1 {
					args = append(args, "0")
				}
				ctx["db"] = args[0]
				conn.SetContext(ctx)
				conn.WriteString("OK")
				return
			}

			// set the default db if there is no db selected
			if ctx["db"] == nil || ctx["db"].(string) == "" {
				ctx["db"] = "0"
			}

			// initialize the selected db
			db, err := selectDB(ctx["db"].(string))
			if err != nil {
				conn.WriteError(fmt.Sprintf("db error: %s", err.Error()))
				return
			}

			// our internal change log
			if changelog.Subscribers(defaultPubSubAllTopic) > 0 {
				changelog.Broadcast(Change{
					Namespace: ctx["db"].(string),
					Command:   todo,
					Arguments: args,
				}, defaultPubSubAllTopic)
			}

			// internal ping-pong
			if todo == "ping" {
				conn.WriteString("PONG")
				return
			}

			// close the connection
			if todo == "quit" {
				conn.WriteString("OK")
				conn.Close()
				return
			}

			// find the required command in our registry
			fn := commands[todo]
			if nil == fn {
				conn.WriteError(fmt.Sprintf("unknown commands [%s]", todo))
				return
			}

			// dispatch the command and catch its errors
			fn(Context{
				Conn:   conn,
				action: todo,
				args:   args,
				db:     db,
			})
		},
		func(conn redcon.Conn) bool {
			conn.SetContext(map[string]interface{}{})
			return true
		},
		nil,
	)
}
