package db

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"sync"

	"github.com/alash3al/redix/internals/version"
)

type Manager struct {
	databases   map[string]*DB
	basedir     string
	defaultOpts Options

	sync.RWMutex
}

func NewManager(basedir string, defaultOpts Options) *Manager {
	m := &Manager{
		databases:   map[string]*DB{},
		basedir:     basedir,
		defaultOpts: defaultOpts,
	}

	m.preopen()

	return m
}

func (m *Manager) OpenDB(dbname string) (*DB, error) {
	m.Lock()
	defer m.Unlock()

	if db, ok := m.databases[dbname]; ok {
		return db, nil
	}

	fullpath := filepath.Join(m.basedir, version.DataLayoutVersion, dbname)

	db, err := newdb(fullpath, &m.defaultOpts)
	if err != nil {
		return nil, err
	}

	m.databases[dbname] = db

	return db, nil
}

func (m *Manager) preopen() {
	datadir := filepath.Join(m.basedir, version.DataLayoutVersion)
	dirs, _ := ioutil.ReadDir(datadir)

	for _, f := range dirs {
		if !f.IsDir() {
			continue
		}

		name := filepath.Base(f.Name())

		_, err := m.OpenDB(name)
		if err != nil {
			log.Fatal(err.Error())
			continue
		}
	}
}

func (m *Manager) CloseAll() {
	m.Lock()
	defer m.Unlock()

	for k, v := range m.databases {
		v.Close()
		delete(m.databases, k)
	}
}
