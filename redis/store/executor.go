package store

import (
	"github.com/alash3al/redix/redis/context"
)

type Executer interface {
	Exec(context.Context) interface{}
}
