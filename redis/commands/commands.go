// Package commands is the commands container
package commands

import (
	"fmt"

	"github.com/alash3al/redix/redis/context"
	"github.com/alash3al/redix/redis/store"
)

type Request struct {
	*context.Context
	store.Store
}

type HandlerFunc func(Request) (interface{}, error)

var (
	Commands = map[string]HandlerFunc{
		"set": set,
	}

	ErrInvalidArgumentsNumber = fmt.Errorf("invalid arguments count specified")
)
