// Package configparser provides a set of utilities for parsing and populating configs
package configparser

import (
	"io/ioutil"
	"os"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

type Config struct {
	Storage struct {
		Driver DatabaseEngine `hcl:"driver"`

		Connection struct {
			Default string `hcl:"default"`
			Cluster struct {
				Read  []string `hcl:"read"`
				Write []string `hcl:"write"`
			} `hcl:"cluster,block"`
		} `hcl:"connection,block"`
	} `hcl:"storage,block"`

	Modules []string `hcl:"modules"`

	Server struct {
		Redis struct {
			ListenAddr string `hcl:"listen"`
		} `hcl:"redis,block"`
	} `hcl:"server,block"`
}

type DatabaseEngine string

const (
	DatabaseEnginePostgres DatabaseEngine = "postgres"
)

func Parse(filename string) (*Config, error) {
	var config Config

	// TODO: use os.ExpandEnvVars in the config filename
	fileContents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	fileContentsExpanded := os.ExpandEnv(string(fileContents))

	err = hclsimple.Decode(filename, []byte(fileContentsExpanded), nil, &config)

	if config.Storage.Connection.Default != "" {
		config.Storage.Connection.Cluster.Read = append(config.Storage.Connection.Cluster.Read, config.Storage.Connection.Default)
		config.Storage.Connection.Cluster.Write = append(config.Storage.Connection.Cluster.Write, config.Storage.Connection.Default)
	}

	return &config, err
}
