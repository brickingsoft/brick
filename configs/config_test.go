package configs_test

import (
	"testing"

	"github.com/brickingsoft/brick/configs"
	"github.com/goccy/go-yaml"
)

func TestMarshal(t *testing.T) {
	type Config struct {
		Name    string         `yaml:"name"`
		Options configs.Config `yaml:"options"`
	}

	type Options struct {
		N int    `yaml:"n"`
		S string `yaml:"s"`
	}

	var b = []byte(`name: test
options:
 s: "s"
 n: 1`)

	config := Config{}
	err := yaml.Unmarshal(b, &config)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("name:", config.Name)

	options := Options{}
	err = config.Options.As(&options)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("options:", options)
}
