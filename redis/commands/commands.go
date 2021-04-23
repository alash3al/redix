// Package commands is the commands container
package commands

import (
	"github.com/alash3al/redix/redis/context"
	"github.com/alash3al/redix/redis/store"
)

type Request struct {
	Context *context.Context
	Store   store.Store
}

type HandlerFunc func(Request) (interface{}, error)

var (
	Commands = map[string]HandlerFunc{}
)
