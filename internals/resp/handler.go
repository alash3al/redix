package resp

// Handlers a registry for available commands
var Handlers = map[string]HandlerFunc{}

// HandlerFunc a command handler
type HandlerFunc func(*Context)
