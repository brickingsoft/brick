package config

import (
	"fmt"

	"github.com/goccy/go-yaml"
)

type Config struct {
}

func (cfg *Config) Path(p string) {
	path := yaml.Path{}
	fmt.Println(path.String())
}
