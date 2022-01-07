package commands

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	"github.com/alash3al/redix/internals/config"
	"github.com/alash3al/redix/internals/datastore/contract"
	"github.com/tidwall/redcon"
)

// Context represents the command context
type Context struct {
	Conn   redcon.Conn
	Engine contract.Engine
	Cfg    *config.Config
	Argv   [][]byte
	Argc   int

	sync.RWMutex
}

// Session fetches the current session map
func (c *Context) Session() map[string]interface{} {
	c.RLock()
	m := c.Conn.Context().(map[string]interface{})
	c.RUnlock()

	return m
}

// SessionSet set a k-v into the current session
func (c *Context) SessionSet(k string, v interface{}) {
	c.Lock()

	m := c.Conn.Context().(map[string]interface{})
	m[k] = v
	c.Conn.SetContext(m)

	c.Unlock()
}

// SessionGet fetches a value from the current session
func (c *Context) SessionGet(k string) (interface{}, bool) {
	val, ok := c.Session()[k]

	return val, ok
}

// AbsoluteKeyPath returns the full key path relative to the namespace the namespace
func (c *Context) AbsoluteKeyPath(k ...[]byte) []byte {
	ns, _ := c.SessionGet("namespace")
	return []byte(ns.(string) + string(bytes.Join(k, []byte("/"))))
}

// Handler a command handler func
type Handler func(*Context)

var (
	commandsMap     = map[string]Handler{}
	commandsMapLock = &sync.RWMutex{}
)

// HandleFunc reigtser a command handler
func HandleFunc(name string, fn Handler) {
	commandsMapLock.Lock()
	defer commandsMapLock.Unlock()

	name = strings.ToLower(name)

	if _, exists := commandsMap[name]; exists {
		panic(fmt.Errorf("command '%s' already exists", name))
	}

	commandsMap[name] = fn
}

// Call executes the specified command name if exists
func Call(name string, ctx *Context) {
	commandsMapLock.RLock()

	name = strings.ToLower(name)

	cmd, exists := commandsMap[name]
	if !exists {
		commandsMapLock.RUnlock()
		ctx.Conn.WriteError(fmt.Sprintf("Err unknown command %s", name))
		return
	}

	commandsMapLock.RUnlock()

	cmd(ctx)
}
