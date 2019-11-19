package server

// Handlers a registry for available commands
var Handlers = map[string]Handler{}

// HandlerFunc a command handler
type HandlerFunc func(Context) error

// Handler represents a command handler
type Handler struct {
	Title       string
	Description string
	Writer      bool
	Reader      bool
	Func        HandlerFunc
}
