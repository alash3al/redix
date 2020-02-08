package resp

// Handlers a registry for available commands
var Handlers = map[string]Handler{}

// HandlerFunc a command handler
type HandlerFunc func(*Context)

// Handler represents a command handler
type Handler struct {
	Title       string
	Description string
	Examples    []string
	Writer      bool
	Reader      bool
	Group       string
	Callback    HandlerFunc
}
