package db

import (
	"sync"
)

var (
	databases = sync.Map{}
)
