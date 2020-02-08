package driver

// Entry represents a key - value pair
type Entry struct {
	Key         []byte
	Value       []byte
	TTL         int
	WriteMerger func(oldValue []byte, newValue []byte) []byte `msgpack:"-" json:"-"`
}
