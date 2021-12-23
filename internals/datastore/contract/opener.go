package contract

// Opener represents a database Open request
type Opener interface {
	Open(string) (DB, error)
}
