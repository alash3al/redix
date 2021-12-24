package contract

// Engine represents an Engine
type Engine interface {
	Opener
	Putter
	Getter
	Iterator
	Hooker
}
