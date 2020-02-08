package resp

import "errors"

var (
	errNoCommand      = errors.New("no command specified")
	errUnknownCommand = errors.New("unknown command specified")
)
