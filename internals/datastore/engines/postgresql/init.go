package postgresql

import "github.com/alash3al/redix/internals/datastore/contract"

// Global consts
const (
	Name = "postgresql"
)

func init() {
	contract.Register(Name, &Engine{})
}
