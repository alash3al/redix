package store

type TokenManager interface {
	TokenCreate() error
	TokenReset(token string) error
	TokenValidate(token string) (bool, error)
}
