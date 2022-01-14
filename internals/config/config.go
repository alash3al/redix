package config

import (
	"io/ioutil"
	"os"

	"github.com/hashicorp/hcl/v2/hclsimple"
)

// Config represents global configs container
type Config struct {
	Server struct {
		Redis struct {
			ListenAddr  string `hcl:"listen"`
			AsyncWrites bool   `hcl:"async"`
			MaxConns    int64  `hcl:"max_connections"`
		} `hcl:"redis,block"`
	} `hcl:"server,block"`

	Engine struct {
		Driver string `hcl:"driver,label"`
		DSN    string `hcl:"dsn"`
	} `hcl:"engine,block"`
}

// Unmarshal parses the specified filename and load it into memory
func Unmarshal(filename string) (*Config, error) {
	configdata, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	configdata = []byte(os.ExpandEnv(string(configdata)))

	var cfg Config

	if err := hclsimple.Decode(filename, configdata, nil, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
