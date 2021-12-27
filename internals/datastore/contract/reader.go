package contract

// GetInput represents a Get request
type GetInput struct {
	Key []byte
}

// GetOutput represents a Get output
type GetOutput struct {
	Value               []byte
	ExpiresAfterSeconds float64
}

// IteratorFunc function used while iterator
type IteratorFunc func([]byte, []byte) bool

// Reader represents read related actions
type Reader interface {
	Get(*GetInput) (*GetOutput, error)
	ForEach(IteratorFunc) error
}
