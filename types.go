package main

import (
	"github.com/tidwall/redcon"
)

// DB Interface
type DB interface {
	// Open(path string) (*DB, error)
	// Incr(k string, incrby int64) (int64, error)
	Set(k, v string, ttl int) error
	MSet(data map[string]string) error
	Get(k string) (string, error)
	MGet(keys []string) []string
	Del(keys []string) error
	Scan(ScannerOpt ScannerOptions) error
}

// ScannerOptions - represents the options for a scanner
type ScannerOptions struct {
	// from where to start
	Offset string

	// whether to include the value of the offset in the result or not
	IncludeOffset bool

	// the prefix that must be exists in each key in the iteration
	Prefix string

	// fetch the values (true) or this is a key only iteration (false)
	FetchValues bool

	// the handler that handles the incoming data
	Handler func(k, v string) bool
}

// type DB

// CommandHandler - represents a handler for a command
type CommandHandler func(c Context)

// Context - represents a handler context
type Context struct {
	redcon.Conn
	db     DB
	action string
	args   []string
}
