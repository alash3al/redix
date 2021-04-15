// Package context provide our main context manager
package context

import "github.com/tidwall/redcon"

type Context struct {
	Conn redcon.Conn

	CurrentToken    string
	CurrentDatabase int
	IsAuthorized    bool
	Command         []byte
	Args            [][]byte
}
