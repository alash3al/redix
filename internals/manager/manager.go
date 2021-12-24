package manager

import (
	"fmt"
	"os"

	"github.com/alash3al/redix/internals/binlog"
	"github.com/alash3al/redix/internals/datastore/contract"
)

// Manager represents a datasource/engines manager
type Manager struct {
	db     contract.Engine
	binlog *binlog.BinLog
	opts   *Options
}

// New initializes a new manager
func New(opts *Options) (*Manager, error) {
	var err error

	if !contract.Exists(opts.DefaultEngine) {
		return nil, fmt.Errorf("unknown specified driver (%s)", opts.DefaultEngine)
	}

	if err = os.MkdirAll(opts.DataDirPath(opts.DefaultEngine), 0755); err != nil {
		return nil, err
	}

	db, err := contract.Open(
		opts.DefaultEngine,
		opts.DataDirPath(opts.DefaultEngine, "redix.data"),
	)

	if err != nil {
		return nil, err
	}

	binlogger, err := binlog.Open(opts.DataDirPath("binlog"))
	if err != nil {
		return nil, err
	}

	mngr := &Manager{
		opts:   opts,
		db:     db,
		binlog: binlogger,
	}

	mngr.db.BeforeCommit(func(input *contract.PutInput) error {
		logEntry := binlog.LogEntry{
			Action:  "Put",
			Payload: input,
		}

		return mngr.binlog.Put(&logEntry)
	})

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

// BinLog returns the binlog handler
func (m *Manager) BinLog() *binlog.BinLog {
	return m.binlog
}
