// Package commands is the commands container
package commands

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/alash3al/redix/redis/ctx"
)

type HandlerFunc func(*ctx.Ctx) (interface{}, error)

var (
	Commands = map[string]HandlerFunc{
		"auth":   auth,
		"select": selectDB,
	}

	ErrInvalidArgumentsNumber = fmt.Errorf("invalid arguments count specified")
)

func auth(c *ctx.Ctx) (interface{}, error) {
	if len(c.Args) < 1 {
		return nil, errors.New("missing authentication token")
	}

	ok, err := c.Store.AuthValidate(c.Args[0])
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, errors.New("invalid authentication token")
	}

	c.CurrentToken = c.Args[0]
	c.IsAuthenticated = true

	return "OK", nil
}

func selectDB(c *ctx.Ctx) (interface{}, error) {
	toBeSelected := 0
	if len(c.Args) > 0 {
		toBeSelected, _ = strconv.Atoi(c.Args[0])
	}

	selectedDB, err := c.Store.Select(c.CurrentToken, toBeSelected)
	if err != nil {
		return nil, err
	}

	c.CurrentDatabase = selectedDB

	return "OK", nil
}
