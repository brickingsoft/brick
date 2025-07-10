// Package main
//
// Usage:
//
// 1. brick: //go:generate go run -mod=mod github.com/brickingsoft/brick/bin/b3 -v .
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {

	ctx := context.Background()

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	cmd := &cli.Command{
		Name:                  "b3",
		Usage:                 "brick cli application",
		Description:           "brick by brick",
		Version:               "v0.0.1",
		EnableShellCompletion: true,
		Commands: []*cli.Command{
			{
				Name:  "create",
				Usage: "create a brick project",
				Action: func(ctx context.Context, c *cli.Command) error {

					return nil
				},
			},
			{
				Name:  "image",
				Usage: "create Dockerfile",
				Action: func(ctx context.Context, c *cli.Command) error {
					return nil
				},
			},
			{
				Name:      "brick",
				Usage:     "scan and generate code.",
				UsageText: "use `//go:generate go run -mod=mod github.com/brickingsoft/brick/bin/b3 -v .` is well.",
				Action: func(ctx context.Context, c *cli.Command) error {
					return nil
				},
			},
		},
	}

	if err := cmd.Run(ctx, os.Args); err != nil {
		log.Error("b3 failed !!!")
	}

}
