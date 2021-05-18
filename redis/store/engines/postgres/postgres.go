// Package postgres represents a storage adapter
package postgres

import (
	"context"

	"github.com/alash3al/redix/configparser"
	"github.com/alash3al/redix/redis/store"
	"github.com/alash3al/redix/utils/roundrobin"
	"github.com/jmoiron/sqlx"
	"github.com/rs/xid"

	_ "embed"

	"github.com/jackc/pgx/v4/pgxpool"
)

//go:embed schema.sql
var schemaSQL string

type KeyType string

const (
	KEYTYPE_STRING KeyType = "str"
	KEYTYPE_INT    KeyType = "int"
	KEYTYPE_FLOAT  KeyType = "float"
)

type Store struct {
	config    *configparser.Config
	readConn  *roundrobin.RR
	writeConn *roundrobin.RR
}

func (s *Store) Connect(config *configparser.Config) (store.Store, error) {
	s.config = config
	s.readConn = roundrobin.New([]interface{}{})
	s.writeConn = roundrobin.New([]interface{}{})

	for _, dsn := range config.Storage.Connection.Cluster.Read {
		conn, err := pgxpool.Connect(context.Background(), dsn)
		if err != nil {
			return nil, err
		}

		s.readConn.Add(conn)
	}

	for _, dsn := range config.Storage.Connection.Cluster.Write {
		conn, err := pgxpool.Connect(context.Background(), dsn)
		if err != nil {
			return nil, err
		}

		s.writeConn.Add(conn)
	}

	if _, err := s.Writer().Exec(context.Background(), schemaSQL); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Store) AuthCreate() (string, error) {
	secret := xid.New().String()

	var id string

	err := s.Writer().QueryRow(
		context.Background(),
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

	err = s.Reader().QueryRow(
		context.Background(),
		`select id, secret from redix_users where id = $1 and secret = $2`,
		inputID,
		inputSecret,
	).Scan(&user.ID, &user.Secret)

	if err != nil {
		return "", err
	}

	user.Secret = xid.New().String()

	_, err = s.Writer().Exec(
		context.Background(),
		`update redix_users set secret = $1 where id = $2`,
		user.Secret,
		user.ID,
	)
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

	err = s.Reader().QueryRow(
		context.Background(),
		`select exists(select * from redix_users where id = $1 and secret = $2)`,
		id,
		secret,
	).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists, nil
}

func (s *Store) Select(token string, db int) (int, error) {
	userID, _, err := parseToken(token)
	if err != nil {
		return -1, err
	}

	err = s.Writer().QueryRow(
		context.Background(),
		`
		insert into redix_databases (user_id, name)
		values($1, $2)
		on conflict (user_id, name) do update set name = excluded.name returning id;
	`, userID, db).Scan(&db)

	if err != nil {
		return -1, err
	}

	return db, nil
}

func (s *Store) Writer() *pgxpool.Conn {
	return s.writeConn.Next().(*pgxpool.Conn)
}

func (s *Store) Reader() *pgxpool.Conn {
	return s.readConn.Next().(*pgxpool.Conn)
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
