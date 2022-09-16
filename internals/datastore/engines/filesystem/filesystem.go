//go:build linux || darwin

package filesystem

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/alash3al/redix/internals/datastore/contract"
)

// Engine represents the contract.Engine implementation
type Engine struct {
	storageDir string
	kvDir      string
}

// Open opens the database
func (e *Engine) Open(dir string) (err error) {
	if err := os.MkdirAll(dir, 0775); err != nil && err != os.ErrExist {
		return err
	}

	if err := os.MkdirAll(filepath.Join(dir, "/kv"), 0775); err != nil && err != os.ErrExist {
		return err
	}

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	e.storageDir = absDir
	e.kvDir = filepath.Join(e.storageDir, "/kv")

	return nil
}

// Write writes into the database
func (e *Engine) Write(input *contract.WriteInput) (*contract.WriteOutput, error) {
	if input == nil {
		return nil, fmt.Errorf("empty input specified")
	}

	if input.Append {
		return nil, fmt.Errorf("(filesystem) unsupported feature (append)")
	}

	if input.Increment {
		return nil, fmt.Errorf("(filesystem) unsupported feature (increment)")
	}

	if input.OnlyIfNotExists {
		return nil, fmt.Errorf("(filesystem) unsupported feature (ifNotExists)")
	}

	if input.TTL > 0 {
		return nil, fmt.Errorf("(filesystem) unsupported feature (TTL)")
	}

	if input.Key == nil {
		if err := os.RemoveAll(e.kvDir); err != nil {
			return nil, err
		}

		return nil, nil
	}

	key := hex.EncodeToString(input.Key)
	keyDataPath := filepath.Join(e.kvDir, key)

	if input.Value == nil {
		if err := os.Remove(keyDataPath); err != nil {
			return nil, err
		}

		return nil, nil
	}

	if err := os.MkdirAll(filepath.Join(e.kvDir, "/kv"), 0775); err != nil && err != os.ErrExist {
		return nil, err
	}

	if _, err := WriteFileWithExclusiveLock(keyDataPath, input.Value); err != nil {
		return nil, err
	}

	return &contract.WriteOutput{
		Value: input.Value,
	}, nil
}

// Get reads from the database
func (e *Engine) Read(input *contract.ReadInput) (*contract.ReadOutput, error) {
	if input == nil {
		return nil, fmt.Errorf("empty input specified")
	}

	key := hex.EncodeToString(input.Key)
	keyDataPath := filepath.Join(e.kvDir, key)

	data, err := ReadFileWithSharedLock(keyDataPath)
	if err != nil {
		if err == os.ErrNotExist {
			return &contract.ReadOutput{}, nil
		}

		return nil, err
	}

	if input.Delete {
		return nil, filepath.WalkDir(e.kvDir, func(path string, d fs.DirEntry, err error) error {
			if path == e.kvDir {
				return nil
			}

			actualKey, err := hex.DecodeString(d.Name())
			if err != nil {
				return err
			}

			if bytes.HasPrefix(actualKey, input.Key) {
				if err := os.Remove(path); err != nil {
					return err
				}
			}

			return nil
		})
	}

	return &contract.ReadOutput{
		Key:    input.Key,
		Value:  data,
		Exists: true,
		TTL:    -1,
	}, nil
}

// Iterate iterates on the whole database stops if the IteratorOpts returns an error
func (e *Engine) Iterate(opts *contract.IteratorOpts) error {
	if opts == nil {
		return fmt.Errorf("empty options specified")
	}

	if opts.Callback == nil {
		return fmt.Errorf("you must specify the callback")
	}

	return filepath.WalkDir(e.kvDir, func(path string, d fs.DirEntry, err error) error {
		if path == e.kvDir {
			return nil
		}

		actualKey, err := hex.DecodeString(d.Name())
		if err != nil {
			return err
		}

		if bytes.HasPrefix(actualKey, opts.Prefix) {
			data, err := ReadFileWithSharedLock(path)
			if err != nil {
				return err
			}

			if err := opts.Callback(&contract.ReadOutput{
				Key:    bytes.TrimPrefix(actualKey, opts.Prefix),
				Value:  data,
				Exists: true,
				TTL:    -1,
			}); err != nil {
				return err
			}
		}

		return nil
	})
}

// Close closes the connection
func (e *Engine) Close() error {
	return nil
}

// Publish not supported in filesystem mode
func (e *Engine) Publish(channel []byte, payload []byte) error {
	return fmt.Errorf("the %s driver doesn't support publish/subscribe", Name)
}

// Subscribe not supported in filesystem mode
func (e *Engine) Subscribe(channel []byte, cb func([]byte) error) error {
	return fmt.Errorf("the %s driver doesn't support publish/subscribe", Name)
}
