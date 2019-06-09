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

type ConfigSetHeader struct {
	Name  string
	Value string
}

type ConfigSetHeaders []ConfigSetHeader

func (cfg ConfigSetHeaders) ToMap() map[string]string {
	r := map[string]string{}
	for _, h := range cfg {
		r[h.Name] = h.Value
	}
	return r
}

type ConfigFrontend struct {
	Listen      string
	Concurrency int
	IsControl   bool `yaml:"is_control"`
	//Rules []ConfigRule
	ReadBufferSize     uint             `yaml:"read_buffer_size"`
	WriteBufferSize    uint             `yaml:"write_buffer_size"`
	MaxRequestBodySize uint             `yaml:"max_request_body_size"`
	MaxConnsPerIP      uint             `yaml:"max_conns_per_ip"`
	SetHeaders         ConfigSetHeaders `yaml:"set_headers"`
}

type ConfigBackend struct {
	Address     string
	URLRegexp   string `yaml:"url_regexp"`
	Connections int
	Return      uint
	SetHeaders  ConfigSetHeaders `yaml:"set_headers"`
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
