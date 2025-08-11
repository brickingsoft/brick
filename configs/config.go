package configs

import (
	"errors"

	"github.com/brickingsoft/brick/pkg/mists"
)

func NewConfig(b []byte) (*Config, error) {
	mist, err := mists.New(b)
	if err != nil {
		return nil, err
	}
	return &Config{mist}, nil
}

type Config struct {
	mist *mists.Mist
}

func (config *Config) As(v any) error {
	return config.mist.Unmarshal(v)
}

func (config *Config) Bytes() []byte {
	return config.mist.Bytes()
}

func (config *Config) Node(name string) (node *Config) {
	if n, exist := config.mist.Node(name); exist {
		node = &Config{mist: n}
		return
	}
	node = &Config{mist: mists.Empty()}
	return
}

func (config *Config) Path(expr string) (node *Config, err error) {
	n, pathErr := config.mist.Path(expr)
	if pathErr != nil {
		node = &Config{mist: mists.Empty()}
		err = pathErr
		return
	}
	node = &Config{mist: n}
	return
}

func (config *Config) Empty() bool {
	return config.mist.Empty()
}

func (config *Config) Merge(target *Config) error {
	return config.mist.Merge(target.mist)
}

func (config *Config) String() string {
	return config.mist.String()
}

func configMistMarshaler(config Config) ([]byte, error) {
	raw := config.Bytes()
	if len(raw) == 0 {
		return []byte{'{', '}'}, nil
	}
	if mists.Valid(raw) {
		return raw, nil
	}
	return nil, errors.New("invalid config bytes")
}

func configMistPtrMarshaler(config *Config) ([]byte, error) {
	if config == nil {
		return []byte{'{', '}'}, nil
	}
	return configMistMarshaler(*config)
}

func configMistUnmarshaler(r *Config, b []byte) error {
	mist, mistErr := mists.New(b)
	if mistErr != nil {
		return mistErr
	}
	r.mist = mist
	return nil
}
