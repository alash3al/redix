package ctx

import (
	"github.com/alash3al/redix/redis/store"
	"github.com/tidwall/redcon"
)

type Ctx struct {
	Conn            redcon.Conn
	Command         string
	Args            []string
	CurrentToken    string
	CurrentDatabase int
	IsAuthenticated bool
	Store           store.Store
}
