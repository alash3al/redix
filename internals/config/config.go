package config

import (
	"io/ioutil"
	"os"

	"github.com/alash3al/redix/internals/manager"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// Config represents global configs container
type Config struct {
	InstanceRole           manager.InstanceRole `envconfig:"INSTANCE_ROLE" required:"true"`
	InstanceRespListenAddr string               `envconfig:"INSTANCE_RESP_LISTEN_ADDR" required:"true"`
	DataDir                string               `envconfig:"DATA_DIR" required:"true"`
	MasterDSN              string               `envconfig:"MASTER_DSN"`
}

// Unmarshal parses the specified filename and load it into memory
func Unmarshal(filename string) (*Config, error) {
	configdata, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	env, err := godotenv.Unmarshal(os.ExpandEnv(string(configdata)))
	if err != nil {
		return nil, err
	}

	for k, v := range env {
		if err := os.Setenv(k, v); err != nil {
			return nil, err
		}
	}

	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
