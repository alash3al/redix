package contract

// Replicator represents replication related actions
type Replicator interface {
	AddReplica(string) error
}
