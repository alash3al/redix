package db

import (
	"errors"
	"sync"
)

// Errors
var (
	ErrDriverNotFound = errors.New("selected database driver isn't available")
)

var (
	databases = sync.Map{}
)
