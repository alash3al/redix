package container

import (
	dbpkg "github.com/alash3al/redix/pkg/db"
)

type Container struct {
	db *dbpkg.DB
}

func NewContainer(db *dbpkg.DB) *Container {
	return &Container{
		db: db,
	}
}
