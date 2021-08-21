// Package configparser provides a set of utilities for parsing and populating configs
package configparser

import (
	"io/ioutil"
	"os"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

type Config struct {
}

func Parse(filename string) (*Config, error) {
	var config Config

	// TODO: use os.ExpandEnvVars in the config filename
	fileContents, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	fileContentsExpanded := os.ExpandEnv(string(fileContents))

	err = hclsimple.Decode(filename, []byte(fileContentsExpanded), nil, &config)

	return &config, err
}
