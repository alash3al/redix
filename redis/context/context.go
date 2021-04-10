// Package context provide our main context manager
package context

type Context struct {
	CurrentToken    string
	CurrentDatabase int
	IsAuthorized    bool
	Command         []byte
	Args            [][]byte
}
