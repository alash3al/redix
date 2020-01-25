package driver

import "errors"

// error related variables
var (
	ErrDriverAlreadyExists = errors.New("the specified driver name is already exisrs")
	ErrDriverNotFound      = errors.New("the requested driver isn't found")
)
