package server

import (
	"gopkg.in/yaml.v2"
	"io"
)

/*type ConfigRule struct {
	ForceBackend string
	ForceRedirect string
}*/

type Logger interface {
	Printf(string, ...interface{})
}

type ConfigFrontend struct {
	Listen      string
	Concurrency int
	//Rules []ConfigRule
}

type ConfigBackend struct {
	Address     string
	URLRegexp   string `yaml:"url_regexp"`
	Connections int
}

type Config struct {
	Frontends []ConfigFrontend
	Backends  []ConfigBackend
}

func NewConfig() *Config {
	return &Config{}
}

func (cfg *Config) Parse(reader io.Reader) error {
	decoder := yaml.NewDecoder(reader)
	return decoder.Decode(cfg)
}
