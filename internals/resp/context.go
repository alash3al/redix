package resp

import (
	"github.com/alash3al/redix/internals/db"
	"github.com/tidwall/redcon"
)

// Context represents a command context
type Context struct {
	conn       redcon.Conn
	args       [][]byte
	db         *db.DB
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
func (c Context) DB() *db.DB {
	return c.db
}

// ChangeDB opens the specified database and set it into the context
func (c Context) ChangeDB(dbname string) (*db.DB, error) {
	db, err := c.serverOpts.Openner(dbname)

	if err != nil {
		return nil, err
	}

	c.db = db

	return db, nil
}
