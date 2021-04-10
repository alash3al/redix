// Package postgres represents a storage adapter
package postgres

import (
	"errors"

	"github.com/alash3al/redix/configparser"
	"github.com/alash3al/redix/redis/context"
	"github.com/alash3al/redix/redis/store"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Store struct {
	config    *configparser.Config
	readConn  []*sqlx.DB
	writeConn []*sqlx.DB
}

func (s *Store) Connect(config *configparser.Config) (store.Store, error) {
	newStore := &Store{}

	newStore.config = config
	newStore.readConn = []*sqlx.DB{}
	newStore.writeConn = []*sqlx.DB{}

	for _, dsn := range config.Connections.Read {
		conn, err := sqlx.Connect("postgres", dsn)
		if err != nil {
			return nil, err
		}

		newStore.readConn = append(s.readConn, conn)
	}

	for _, dsn := range config.Connections.Write {
		conn, err := sqlx.Connect("postgres", dsn)
		if err != nil {
			return nil, err
		}

		newStore.writeConn = append(s.writeConn, conn)
	}

	return newStore, nil
}

func (s *Store) TokenCreate() error {
	return nil
}

func (s *Store) TokenReset(token string) error {
	return nil
}

func (s *Store) TokenValidate(token string) (bool, error) {
	return true, nil
}

func (s *Store) Exec(ctx context.Context) interface{} {
	return errors.New("COMMAND NOT IMPLEMENTED")
}

func (s *Store) Close() error {
	return nil
}
