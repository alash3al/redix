package server

import (
	"github.com/alash3al/redix/db"
	"github.com/tidwall/redcon"
)

// Context represents a command context
type Context struct {
	conn    redcon.Conn
	command redcon.Command
	db      *db.DB
}

// Conn returns the underlying connection
func (c Context) Conn() redcon.Conn {
	return c.conn
}

// Command returns the incomming commqand
func (c Context) Command() *redcon.Command {
	return &c.command
}

// DB returns the underlying database instance
func (c Context) DB() *db.DB {
	return c.db
}
