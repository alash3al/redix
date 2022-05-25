package filesystem

import "github.com/alash3al/redix/internals/datastore/contract"

// Global consts
const (
	Name = "filesystem"
)

func init() {
	contract.Register(Name, &Engine{})
}
