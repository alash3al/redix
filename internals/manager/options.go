package manager

import (
	"path/filepath"
)

// InstanceRole represents the role of the initiated manager instance
type InstanceRole string

// Available instance roles
const (
	InstanceRoleMaster  InstanceRole = "master"
	InstanceRoleReplica InstanceRole = "replica"
)

// Options manager options
type Options struct {
	DataDir           string
	DefaultEngine     string
	InstanceRole      InstanceRole
	MasterRESPDSN     string
	MasterHTTPBaseURL string
	ReplicasDSN       []string
	MaxWalSize        string
	RESPListenAddr    string
}

// DataDirPath returns full path relative to the datadir
func (opts *Options) DataDirPath(elem ...string) string {
	elem = append([]string{opts.DataDir}, elem...)
	return filepath.Join(elem...)
}
