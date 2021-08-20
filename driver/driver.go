package driver

type Driver interface {
	Open() (Driver, error)
	Close() error
	Put(*RawValueEntry) error
	Get(key string) (*RawValueEntry, error)
	Exists(key string) (bool, error)
	Walk(func(*RawValueEntry) bool) error
	WalkPrefix(string, func(*RawValueEntry) bool) error
}
