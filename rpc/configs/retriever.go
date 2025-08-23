package configs

import (
	"context"
	"embed"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/brickingsoft/brick/pkg/fs/mem"
)

type Retriever interface {
	Retrieve(ctx context.Context) (config *Config, err error)
}

type RetrieverOptions struct {
	Dir fs.FS
}

type RetrieverOption func(*RetrieverOptions) error

func WithRetrieverDir(dir string) RetrieverOption {
	return func(options *RetrieverOptions) (err error) {
		dir = strings.TrimSpace(dir)
		dir, err = filepath.Abs(dir)
		if err != nil {
			return
		}
		dir = filepath.ToSlash(filepath.Clean(dir))
		options.Dir = os.DirFS(dir)
		return
	}
}

func WithRetrieverEmbedDir(dir *embed.FS) RetrieverOption {
	return func(options *RetrieverOptions) error {
		if dir == nil {
			return errors.New("dir is nil")
		}
		entries, readErr := dir.ReadDir(".")
		if readErr != nil {
			return readErr
		}
		if len(entries) != 1 {
			return errors.New("expected only one directory")
		}

		sub, subErr := fs.Sub(dir, entries[0].Name())
		if subErr != nil {
			return subErr
		}
		options.Dir = sub
		return nil
	}
}

func WithRetrieverMemDir(dir *mem.Dir) RetrieverOption {
	return func(options *RetrieverOptions) error {
		if dir == nil {
			return errors.New("dir is nil")
		}
		options.Dir = dir
		return nil
	}
}

func MultiLevelRetriever(options ...RetrieverOption) Retriever {
	opts := &RetrieverOptions{
		Dir: os.DirFS("configs.d"),
	}
	for _, option := range options {
		if err := option(opts); err != nil {
			return &multiLevelConfigRetriever{
				err: err,
			}
		}
	}
	if opts.Dir == nil {
		return &multiLevelConfigRetriever{
			err: errors.New("dir is nil"),
		}
	}
	return &multiLevelConfigRetriever{
		dir: opts.Dir,
		err: nil,
	}
}

type multiLevelConfigRetriever struct {
	dir fs.FS
	err error
}

func (retriever *multiLevelConfigRetriever) Retrieve(_ context.Context) (config *Config, err error) {
	if retriever.err != nil {
		err = errors.Join(errors.New("retriever config failed"), retriever.err)
		return
	}
	if retriever.dir == nil {
		err = errors.Join(errors.New("retriever config failed"), errors.New("dir is nil"))
		return
	}
	entries, readErr := fs.ReadDir(retriever.dir, ".")
	if readErr != nil {
		err = errors.Join(errors.New("retriever config failed"), readErr)
		return
	}
	if len(entries) == 0 {
		err = errors.Join(errors.New("retriever config failed"), errors.New("config dir is empty"))
		return
	}
	exp, expErr := regexp.Compile(`app(?:\.[a-zA-Z]+)?\.yaml`)
	if expErr != nil {
		err = errors.Join(errors.New("retriever config failed"), expErr)
		return
	}
	active, activeErr := retriever.active()
	if activeErr != nil {
		err = errors.Join(errors.New("retriever config failed"), activeErr)
		return
	}
	activeName := ""
	if active != "" {
		activeName = fmt.Sprintf("app.%s.yaml", active)
	}
	var baseFile fs.DirEntry
	var activeFile fs.DirEntry
	for _, entry := range entries {
		name := entry.Name()
		if exp.MatchString(name) {
			if name == "app.yaml" {
				baseFile = entry
				continue
			}
			if activeName != "" && activeName == name {
				activeFile = entry
			}
		}
	}
	if baseFile == nil && activeFile == nil {
		err = errors.New("no config file in dir")
		return
	}
	var (
		baseConfig   *Config
		activeConfig *Config
	)
	if baseFile != nil {
		baseConfig, err = retriever.readConfig(baseFile.Name())
		if err != nil {
			err = errors.Join(errors.New("retriever config failed"), err)
			return
		}
	}
	if activeFile != nil {
		activeConfig, err = retriever.readConfig(activeFile.Name())
		if err != nil {
			err = errors.Join(errors.New("retriever config failed"), err)
			return
		}
	}

	if baseConfig != nil && activeConfig != nil {
		if err = baseConfig.Merge(activeConfig); err != nil {
			return
		}
		config = baseConfig
		return
	}

	if baseConfig != nil && activeConfig == nil {
		config = baseConfig
		return
	}

	config = activeConfig
	return
}

func (retriever *multiLevelConfigRetriever) readConfig(name string) (config *Config, err error) {
	b, readErr := fs.ReadFile(retriever.dir, name)
	if readErr != nil {
		err = readErr
		return
	}
	config, err = NewConfig(b)
	return
}

func (retriever *multiLevelConfigRetriever) active() (active string, err error) {
	active = os.Getenv("BRICK_ACTIVE")
	active = strings.TrimSpace(active)
	if active != "" {
		return
	}
	var flags *flag.FlagSet
	if len(os.Args) == 0 {
		flags = flag.NewFlagSet("", flag.ContinueOnError)
	} else {
		flags = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	}
	flags.StringVar(&active, "active", "", "active config file")
	if err = flags.Parse(os.Args[1:]); err != nil {
		return
	}
	active = strings.TrimSpace(active)
	return
}
