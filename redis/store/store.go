// Package store provider the contract for each store adapter
package store

type Store interface {
	Connector
	Auth
	Executer
}
