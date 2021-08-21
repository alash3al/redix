package memory

import "github.com/alash3al/redix/internals/driver"

const Name = "memory"

func init() {
	driver.Register(Name, &Engine{})
}
