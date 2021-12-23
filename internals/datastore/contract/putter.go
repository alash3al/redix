package contract

import "time"

// PutInput represents a PUT request
type PutInput struct {
	Key             []byte
	Value           []byte
	OnlyIfNotExists bool
	TTL             time.Duration
}

// PutOutput represents a PUT output
type PutOutput struct{}

// Putter represents a Put related actions
type Putter interface {
	Put(*PutInput) (*PutOutput, error)
}
