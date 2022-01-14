package commands

import (
	"fmt"
	"strings"
	"sync"
)

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
