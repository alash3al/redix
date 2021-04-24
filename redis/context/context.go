// Package context provide our main context manager
package context

import (
	"strings"

	"github.com/tidwall/redcon"
)

type Context struct {
	redcon.Conn

	CurrentToken    string
	CurrentDatabase string
	IsAuthorized    bool
	Command         []byte
	Args            [][]byte
}

func (c Context) ArgsIndexedTable() (m map[string]int) {
	m = map[string]int{}

	for index, arg := range c.Args {
		m[strings.ToLower(string(arg))] = index
	}

	return
}

func (c Context) ArgsIndexExists(i int) bool {
	return i <= len(c.Args)-1
}
