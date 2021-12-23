package manager

import (
	"fmt"
	"os"

	"github.com/alash3al/redix/internals/datastore/contract"
)

// Manager represents a datasource/engines manager
type Manager struct {
	db   contract.Engine
	opts *Options
}

// New initializes a new manager
func New(opts *Options) (*Manager, error) {
	var err error

	if err = os.MkdirAll(opts.DatabasesPath(), 0755); err != nil {
		return nil, err
	}

	if !contract.Exists(opts.DefaultEngine) {
		return nil, fmt.Errorf("unknown specified driver (%s)", opts.DefaultEngine)
	}

	db, err := contract.Open(
		opts.DefaultEngine,
		opts.DatabasesPath("redix.data"),
	)

	if err != nil {
		return nil, err
	}

	mngr := &Manager{
		opts: opts,
		db:   db,
	}

	return mngr, nil
}

// Put puts the specified input into the specified dbIndex
func (m *Manager) Put(input *contract.PutInput) (*contract.PutOutput, error) {
	return m.db.Put(input)
}

// Get fetches the specified input into the specified dbIndex
func (m *Manager) Get(input *contract.GetInput) (*contract.GetOutput, error) {
	return m.db.Get(input)
}

// ForEach iterate over each key-value in the store using the fn, when fn returns false means break the loop
func (m *Manager) ForEach(fn contract.IteratorFunc) error {
	return m.db.ForEach(fn)
}
