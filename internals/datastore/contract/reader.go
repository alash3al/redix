package contract

import "io"

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

// ExportFunc a function used as callback for exporting purposes
// first is the database size, you should return a writer where the dump will be written
type ExportFunc func(int64) io.Writer

// Reader represents read related actions
type Reader interface {
	Get(*GetInput) (*GetOutput, error)
	ForEach(IteratorFunc) error
	Export(ExportFunc) error
}
