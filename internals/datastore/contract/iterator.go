package contract

// IteratorFunc function used while iterator
type IteratorFunc func([]byte, []byte) bool

// Iterator represents iteration related actions
type Iterator interface {
	ForEach(IteratorFunc) error
}
