package postgresql

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/alash3al/redix/internals/datastore/contract"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Engine represents the contract.Engine implementation
type Engine struct {
	conn *pgxpool.Pool
}

// Open opens the database
func (e *Engine) Open(dsn string) (err error) {
	e.conn, err = pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		return err
	}

	if _, err := e.conn.Exec(
		context.Background(),
		`
			CREATE EXTENSION IF NOT EXISTS pg_trgm;

			CREATE TABLE IF NOT EXISTS redix_data_v5 (
				_id 		BIGSERIAL PRIMARY KEY,
				_expires_at BIGINT,
				_key 		TEXT,
				_value 		JSONB
			);

			CREATE UNIQUE INDEX IF NOT EXISTS uniq_idx_redix_data_v5_key ON redix_data_v5 (_key);

			CREATE INDEX IF NOT EXISTS trgm_idx_redix_data_v5_key ON redix_data_v5 USING GIN(_key gin_trgm_ops);

			CREATE INDEX IF NOT EXISTS idx_redix_data_v5_expires_at ON redix_data_v5 (_expires_at);
		`,
	); err != nil {
		return err
	}

	go (func() {
		for {
			now := time.Now().UnixNano()

			if _, err := e.conn.Exec(
				context.Background(),
				`DELETE FROM redix_data_v5 WHERE _expires_at != 0 and _expires_at <= $1`,
				now,
			); err != nil {
				panic(err)
			}

			time.Sleep(time.Second * 1)
		}
	})()

	return nil
}

// Write writes into the database
func (e *Engine) Write(input *contract.WriteInput) (*contract.WriteOutput, error) {
	if input == nil {
		return nil, fmt.Errorf("empty input specified")
	}

	if input.Value == nil {
		if _, err := e.conn.Exec(context.Background(), "DELETE FROM redix_data_v5 WHERE _key LIKE $1", append(input.Key, '%')); err != nil {
			return nil, err
		}

		return nil, nil
	}

	insertQuery := []string{"INSERT INTO redix_data_v5(_key, _value, _expires_at) VALUES($1, $2, $3)"}
	appending := false
	isNumber := false
	ttl := int64(0)

	var val interface{} = string(input.Value)

	if input.TTL > 0 {
		ttl = time.Now().Add(input.TTL).UnixNano()
	}

	if fval, err := strconv.ParseFloat(string(input.Value), 64); err == nil {
		isNumber = true
		val = fval
	}

	if input.OnlyIfNotExists {
		insertQuery = append(insertQuery, "ON CONFLICT (_key) DO NOTHING")
	} else if input.Increment {
		if !isNumber {
			return nil, fmt.Errorf("the specified value is not a number")
		}

		appending = true
		insertQuery = append(insertQuery, "ON CONFLICT (_key) DO UPDATE SET _value = (EXCLUDED._value::text::float + redix_data_v5._value::text::float)::text::jsonb")
	} else if input.Append {
		appending = true
		insertQuery = append(insertQuery, "ON CONFLICT (_key) DO UPDATE SET _value = (redix_data_v5._value::text || EXCLUDED._value::text)::jsonb")
	} else {
		appending = true
		insertQuery = append(insertQuery, "ON CONFLICT (_key) DO UPDATE SET _value = $2::jsonb")
	}

	if appending && !input.KeepTTL {
		insertQuery = append(insertQuery, ", _expires_at = $3::bigint")
	}

	insertQuery = append(insertQuery, "RETURNING _value, _expires_at")

	var retVal []byte
	var retExpiresAt int64

	jsonVal, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}

	if err := e.conn.QueryRow(
		context.Background(),
		strings.Join(insertQuery, " "),
		input.Key, string(jsonVal), ttl,
	).Scan(&retVal, &retExpiresAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &contract.WriteOutput{
		Value: retVal,
		TTL:   time.Now().Sub(time.Unix(0, retExpiresAt)),
	}, nil
}

// Get reads from the database
func (e *Engine) Read(input *contract.ReadInput) (*contract.ReadOutput, error) {
	if input == nil {
		return nil, fmt.Errorf("empty input specified")
	}

	var retQueryVal []byte
	var retVal interface{}
	var retExpiresAt int64

	if err := e.conn.QueryRow(
		context.Background(),
		"SELECT _value, _expires_at FROM redix_data_v5 WHERE _key = $1",
		input.Key,
	).Scan(&retQueryVal, &retExpiresAt); err != nil {
		if err == pgx.ErrNoRows {
			return &contract.ReadOutput{}, nil
		}

		return nil, err
	}

	if err := json.Unmarshal(retQueryVal, &retVal); err != nil {
		return nil, err
	}

	readOutput := contract.ReadOutput{
		Key:    input.Key,
		Value:  []byte(fmt.Sprintf("%v", retVal)),
		TTL:    0,
		Exists: true,
	}

	if retExpiresAt != 0 {
		readOutput.TTL = time.Unix(0, retExpiresAt).Sub(time.Now())
	}

	if readOutput.TTL < 0 {
		return &contract.ReadOutput{}, nil
	}

	return &readOutput, nil
}

// Iterate iterates on the whole database stops if the IteratorOpts returns an error
func (e *Engine) Iterate(opts *contract.IteratorOpts) error {
	if opts == nil {
		return fmt.Errorf("empty options specified")
	}

	if opts.Callback == nil {
		return fmt.Errorf("you must specify the callback")
	}

	iter, err := e.conn.Query(context.Background(), "SELECT _key, _value, _expires_at FROM redix_data_v5 WHERE _key LIKE $1 ORDER BY _id ASC", append(opts.Prefix, '%'))
	if err != nil {
		return err
	}
	defer iter.Close()

	for iter.Next() {
		var key, value []byte
		var expiresAt int64

		if err := iter.Scan(&key, &value, &expiresAt); err != nil {
			return err
		}

		var parsedVal interface{}

		if err := json.Unmarshal(value, &parsedVal); err != nil {
			return err
		}

		readOutput := contract.ReadOutput{
			Key:   key,
			Value: []byte(fmt.Sprintf("%v", parsedVal)),
			TTL:   0,
		}

		if expiresAt != 0 {
			readOutput.TTL = time.Unix(0, expiresAt).Sub(time.Now())
		}

		// expired
		if readOutput.TTL < 0 {
			continue
		}

		if err := opts.Callback(&readOutput); err != nil {
			return err
		}
	}

	return iter.Err()
}

// Close closes the connection
func (e *Engine) Close() error {
	e.conn.Close()
	return nil
}
