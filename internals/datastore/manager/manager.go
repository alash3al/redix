package manager

import (
	"fmt"
	"os"
	"sync"

	"github.com/alash3al/redix/internals/datastore/contract"
)

// Manager represents a datasource/engines manager
type Manager struct {
	databases     map[int]contract.Engine
	statemachine  contract.Engine
	databasesLock *sync.RWMutex
	opts          *Options
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

	mngr := &Manager{
		databases:     make(map[int]contract.Engine),
		databasesLock: new(sync.RWMutex),
		opts:          opts,
	}

	mngr.statemachine, err = contract.Open(opts.DefaultEngine, opts.StateMachinePath())
	if err != nil {
		return nil, err
	}

	return mngr, nil
}

// Select pick the specified database name for usage
func (m *Manager) Select(name int) (contract.Engine, error) {
	var err error

	m.databasesLock.RLock()

	db, exists := m.databases[name]
	if !exists {
		db, err = contract.Open(m.opts.DefaultEngine, m.opts.DatabasePath(name))
	}

	m.databasesLock.RUnlock()

	if err != nil {
		return nil, err
	}

	m.databasesLock.Lock()
	defer m.databasesLock.Unlock()

	if !exists {
		m.databases[name] = db
	}

	return db, nil
}

// Put puts the specified input into the specified dbIndex
func (m *Manager) Put(dbIndex int, input *contract.PutInput) (*contract.PutOutput, error) {
	db, err := m.Select(dbIndex)
	if err != nil {
		return nil, err
	}

	return db.Put(input)
}

// Delete the specified input from the specified dbIndex
func (m *Manager) Delete(dbIndex int, input *contract.DeleteInput) (*contract.DeleteOutput, error) {
	db, err := m.Select(dbIndex)
	if err != nil {
		return nil, err
	}

	return db.Delete(input)
}

// Get fetches the specified input into the specified dbIndex
func (m *Manager) Get(dbIndex int, input *contract.GetInput) (*contract.GetOutput, error) {
	db, err := m.Select(dbIndex)
	if err != nil {
		return nil, err
	}

	return db.Get(input)
}
