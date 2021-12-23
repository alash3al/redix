package contract

// GetInput represents a Get request
type GetInput struct {
	Key    []byte
	Delete bool
}

// GetOutput represents a Get output
type GetOutput struct {
	Value               []byte
	ExpiresAfterSeconds float64
}

// Getter represents Get related actions
type Getter interface {
	Get(*GetInput) (*GetOutput, error)
}
