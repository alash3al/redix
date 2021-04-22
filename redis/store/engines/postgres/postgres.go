// Package postgres represents a storage adapter
package postgres

import (
	"errors"

	"github.com/alash3al/redix/configparser"
	"github.com/alash3al/redix/redis/context"
	"github.com/alash3al/redix/redis/store"
	"github.com/alash3al/redix/utils/roundrobin"
	"github.com/jmoiron/sqlx"
	"github.com/rs/xid"

	_ "embed"

	_ "github.com/lib/pq"
)

//go:embed schema.sql
var schemaSQL string

type Store struct {
	config    *configparser.Config
	readConn  *roundrobin.RR
	writeConn *roundrobin.RR
}

func (s *Store) Connect(config *configparser.Config) (store.Store, error) {
	newStore := &Store{}

	newStore.config = config
	newStore.readConn = roundrobin.New([]interface{}{})
	newStore.writeConn = roundrobin.New([]interface{}{})

	for _, dsn := range config.Storage.Connection.Cluster.Read {
		conn, err := sqlx.Connect("postgres", dsn)
		if err != nil {
			return nil, err
		}

		newStore.readConn.Add(conn)
	}

	for _, dsn := range config.Storage.Connection.Cluster.Write {
		conn, err := sqlx.Connect("postgres", dsn)
		if err != nil {
			return nil, err
		}

		newStore.writeConn.Add(conn)
	}

	return newStore, newStore.migrate()
}

func (s Store) migrate() error {
	if _, err := s.getWriteConn().Exec(schemaSQL); err != nil {
		return err
	}

	return nil
}

func (s Store) AuthRequired() bool {
	return true
}

func (s *Store) AuthCreate() (string, error) {
	secret := xid.New().String()

	var id string

	err := s.getWriteConn().QueryRowx(
		`INSERT INTO redix_users(secret) values($1) RETURNING id`,
		secret,
	).Scan(&id)

	if err != nil {
		return "", err
	}

	return generateToken(id, secret), nil
}

func (s *Store) AuthReset(token string) (string, error) {
	inputID, inputSecret, err := parseToken(token)
	if err != nil {
		return "", err
	}

	var user struct {
		ID     string `db:"id"`
		Secret string `db:"secret"`
	}

	err = s.getReadConn().QueryRowx(
		`select id, secret from redix_users where id = $1 and secret = $2`,
		inputID,
		inputSecret,
	).StructScan(&user)

	if err != nil {
		return "", err
	}

	user.Secret = xid.New().String()

	_, err = s.getWriteConn().Exec(`update redix_users set secret = $1 where id = $2`, user.Secret, user.ID)
	if err != nil {
		return "", err
	}

	return generateToken(user.ID, user.Secret), nil
}

func (s *Store) AuthValidate(token string) (bool, error) {
	var exists bool

	id, secret, err := parseToken(token)
	if err != nil {
		return false, err
	}

	err = s.getReadConn().QueryRowx(
		`select exists(select * from redix_users where id = $1 and secret = $2)`,
		id,
		secret,
	).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}

func (s *Store) Exec(ctx *context.Context) (interface{}, error) {
	return nil, errors.New("COMMAND NOT IMPLEMENTED")
}

func (s *Store) getWriteConn() *sqlx.DB {
	return s.writeConn.Next().(*sqlx.DB)
}

func (s *Store) getReadConn() *sqlx.DB {
	return s.readConn.Next().(*sqlx.DB)
}

func (s *Store) Close() (err error) {
	for {
		conn := s.readConn.Next().(*sqlx.DB)

		if err = conn.Close(); err != nil {
			break
		}

		s.readConn.Remove(conn)
	}

	return
}
