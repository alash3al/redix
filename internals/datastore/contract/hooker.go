package contract

// BeforeCommitFunc represents the callback that will be executed before calling internal put
type BeforeCommitFunc func(*PutInput) error

// Hooker represents internal hooks related actions
type Hooker interface {
	BeforeCommit(BeforeCommitFunc)
}
