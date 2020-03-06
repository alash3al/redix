package db

import (
	_ "github.com/alash3al/goukv/providers/goleveldb"
)

type Provider = string

const (
	LevelDBProvder Provider = "goleveldb"
)
