package manager

import (
	"fmt"
	"path/filepath"
)

// Options manager options
type Options struct {
	DataDir       string
	DefaultEngine string
}

// DataDirPath returns full path relative to the datadir
func (opts *Options) DataDirPath(elem ...string) string {
	elem = append([]string{opts.DataDir}, elem...)
	return filepath.Join(elem...)
}

// DatabasesPath returns the full database path relative to the datadir and based on
func (opts *Options) DatabasesPath(elem ...string) string {
	elem = append([]string{"databases", opts.DefaultEngine}, elem...)
	return opts.DataDirPath(elem...)
}

// DatabasePath returns the full database path relative to the datadir and based on
func (opts *Options) DatabasePath(name int) string {
	return opts.DatabasesPath(fmt.Sprintf("%d.rdb", name))
}

// StateMachinePath returns the state-machine path
func (opts *Options) StateMachinePath() string {
	return opts.DatabasesPath("statemachine.rdb")
}
