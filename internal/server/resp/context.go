package resp

import (
	containerapi "github.com/alash3al/redix/pkg/db/container"
	"github.com/tidwall/redcon"
)

// Context represents a command context
type Context struct {
	conn       redcon.Conn
	args       [][]byte
	container  *containerapi.Container
	serverOpts Options
}

// Conn returns the underlying connection
func (c Context) Conn() redcon.Conn {
	return c.conn
}

// Args returns the incomming commqand
func (c Context) Args() [][]byte {
	return c.args
}

// DB returns the underlying database instance
func (c Context) Container() *containerapi.Container {
	return c.container
}

// ChangeDB opens the specified database and set it into the context
func (c Context) ChangeDB(dbname string) (*containerapi.Container, error) {
	db, err := c.serverOpts.Openner(dbname)

	if err != nil {
		return nil, err
	}

	c.container = containerapi.NewContainer(db)

	return c.container, nil
}
