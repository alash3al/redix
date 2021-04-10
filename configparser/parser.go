// Package configparser provides a set of utilities for parsing and populating configs
package configparser

import "github.com/hashicorp/hcl/v2/hclsimple"

type Config struct {
	Engine DatabaseEngine `hcl:"engine"`

	Modules []string `hcl:"modules"`

	Server struct {
		Redis struct {
			ListenAddr string `hcl:"listen"`
		} `hcl:"redis,block"`
	} `hcl:"server,block"`

	Connections struct {
		Read  []string `hcl:"read"`
		Write []string `hcl:"write"`
	} `hcl:"connection,block"`
}

type DatabaseEngine string

const (
	DatabaseEnginePostgres DatabaseEngine = "postgres"
)

func Parse(filename string) (*Config, error) {
	var config Config

	err := hclsimple.DecodeFile(filename, nil, &config)

	return &config, err
}
