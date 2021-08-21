package memory

import "github.com/alash3al/redix/driver"

const Name = "memory"

func init() {
	driver.Register(Name, &Engine{})
}
