package contract

// DB represents a DB
type DB interface {
	Opener
	Putter
	Getter
	Deleter
}
