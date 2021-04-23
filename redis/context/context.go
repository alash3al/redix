// Package context provide our main context manager
package context

import "github.com/tidwall/redcon"

type Context struct {
	redcon.Conn

	CurrentToken    string
	CurrentDatabase string
	IsAuthorized    bool
	Command         []byte
	Args            [][]byte
}
