package commands

import (
	"fmt"
	"strconv"
)

func set(req Request) (interface{}, error) {
	if len(req.Args) < 2 {
		return nil, ErrInvalidArgumentsNumber
	}

	key, value := string(req.Args[0]), string(req.Args[1])

	_, _ = key, value

	ex, px, nx, xx := -1, -1, false, false

	var err error

	argsIndex := req.ArgsIndexedTable()

	if index, exists := argsIndex["ex"]; exists {
		if !req.ArgsIndexExists(index + 1) {
			return nil, ErrInvalidArgumentsNumber
		}

		ex, err = strconv.Atoi(string(req.Args[index+1]))
	} else if index, exists := argsIndex["px"]; exists {
		if !req.ArgsIndexExists(index + 1) {
			return nil, ErrInvalidArgumentsNumber
		}

		ex, err = strconv.Atoi(string(req.Args[index+1]))
	}

	if _, exists := argsIndex["nx"]; exists {
		nx = true
	} else if _, exists := argsIndex["xx"]; exists {
		xx = true
	}

	fmt.Println(key, value, ex, px, nx, xx)

	return nil, err
}
