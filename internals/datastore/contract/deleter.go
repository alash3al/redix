package contract

// DeleteInput represents a Delete Request
type DeleteInput struct {
	Key []byte
}

// DeleteOutput represents a Delete response
type DeleteOutput struct{}

// Deleter represents Delete related actions
type Deleter interface {
	Delete(*DeleteInput) (*DeleteOutput, error)
}
