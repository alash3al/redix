package contract

import (
	"io"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

// WriteInput represents a PUT request
type WriteInput struct {
	Key             []byte
	Value           []byte
	Delete          bool
	OnlyIfNotExists bool
	TTL             time.Duration
	KeepTTL         bool
}

// Marshal encodes the current input into msgpack
func (i *WriteInput) Marshal() ([]byte, error) {
	return msgpack.Marshal(i)
}

// UnmarshalWriteInput unmarshal the specified data into a WriteInput
func UnmarshalWriteInput(data []byte) (*WriteInput, error) {
	var wi WriteInput

	if err := msgpack.Unmarshal(data, &wi); err != nil {
		return nil, err
	}

	return &wi, nil
}

// WriteOutput represents a PUT output
type WriteOutput struct{}

// Writer represents a Put related actions
type Writer interface {
	Write(*WriteInput) (*WriteOutput, error)
	Import(io.Reader) (int64, error)
}
