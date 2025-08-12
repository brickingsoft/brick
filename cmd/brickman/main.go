package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/brickingsoft/brick/cmd/brickman/actions/cert"
	"github.com/brickingsoft/brick/cmd/brickman/actions/create"
	"github.com/brickingsoft/brick/cmd/brickman/actions/doc"
	"github.com/brickingsoft/brick/cmd/brickman/actions/dockerfile"
	"github.com/urfave/cli/v3"
)

func main() {
	ctx := context.Background()

	cmd := &cli.Command{
		Name:                  "brickman",
		Usage:                 "brick manager",
		Description:           "brick manager cli application",
		Version:               "v0.0.1",
		EnableShellCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "log",
				Usage:   "switch log level",
				Aliases: []string{"l"},
				Validator: func(s string) error {
					level := strings.ToLower(strings.TrimSpace(s))
					switch level {
					case "debug":
					case "info":
					case "warn":
					case "error":
					default:
						return fmt.Errorf("invalid log level: %s", s)
					}
					return nil
				},
			},
		},
		Commands: []*cli.Command{
			{
				Name:   create.Name,
				Usage:  create.Usage,
				Flags:  create.Flags,
				Action: create.Action,
			},
			{
				Name:   cert.Name,
				Usage:  cert.Usage,
				Flags:  cert.Flags,
				Action: cert.Action,
			},
			{
				Name:   dockerfile.Name,
				Usage:  dockerfile.Usage,
				Flags:  dockerfile.Flags,
				Action: dockerfile.Action,
			},
			{
				Name:   doc.Name,
				Usage:  doc.Usage,
				Flags:  doc.Flags,
				Action: doc.Action,
			},
		},
	}

	if err := cmd.Run(ctx, os.Args); err != nil {
		fmt.Println(err)
	}
}
