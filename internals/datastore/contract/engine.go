package contract

// Engine global consts
const (
	StateMachineLastWriteTimeKeyName = "last_write_name"
)

// Engine represents an Engine
type Engine interface {
	Opener
	Putter
	Getter
	Deleter
	// StateMachineSetter
}
