package main

import (
	"github.com/alash3al/redix/kvstore"
	"github.com/tidwall/redcon"
)

// CommandHandler - represents a handler for a command
type CommandHandler func(c Context)

// Context - represents a handler context
type Context struct {
	redcon.Conn
	db     kvstore.DB
	action string
	args   []string
}

// Change - a change feed
type Change struct {
	Namespace string   `json:"namespace"`
	Command   string   `json:"command"`
	Arguments []string `json:"arguments"`
}
