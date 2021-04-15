package store

type Auth interface {
	AuthCreate() error
	AuthReset(token string) error
	AuthValidate(token string) (bool, error)
}
