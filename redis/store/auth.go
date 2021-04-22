package store

type Auth interface {
	AuthRequired() bool
	AuthCreate() (string, error)
	AuthReset(token string) (string, error)
	AuthValidate(token string) (bool, error)
}
