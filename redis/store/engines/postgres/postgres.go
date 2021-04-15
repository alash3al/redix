// Package postgres represents a storage adapter
package postgres

import (
	"errors"

	"github.com/alash3al/redix/configparser"
	"github.com/alash3al/redix/redis/context"
	"github.com/alash3al/redix/redis/store"
	"github.com/jmoiron/sqlx"

	_ "embed"

	_ "github.com/lib/pq"
)

//go:embed schema.sql
var schemaSQL string

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

	for _, dsn := range config.Storage.Connection.Cluster.Read {
		conn, err := sqlx.Connect("postgres", dsn)
		if err != nil {
			return nil, err
		}

		newStore.readConn = append(s.readConn, conn)
	}

	for _, dsn := range config.Storage.Connection.Cluster.Write {
		conn, err := sqlx.Connect("postgres", dsn)
		if err != nil {
			return nil, err
		}

		newStore.writeConn = append(s.writeConn, conn)
	}

	return newStore, newStore.migrate()
}

func (s Store) migrate() error {
	for _, conn := range s.writeConn {
		if _, err := conn.Exec(schemaSQL); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) AuthCreate() error {
	return nil
}

func (s *Store) AuthReset(token string) error {
	return nil
}

func (s *Store) AuthValidate(token string) (bool, error) {
	return true, nil
}

func (s *Store) Exec(ctx context.Context) (interface{}, error) {
	return nil, errors.New("COMMAND NOT IMPLEMENTED")
}

func (s *Store) Close() error {
	return nil
}
