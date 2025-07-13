package main

import (
	"context"
	"fmt"
	"os"

	"github.com/brickingsoft/brick/cmd/brick/create"
	"github.com/brickingsoft/brick/cmd/brick/doc"
	"github.com/brickingsoft/brick/cmd/brick/dockerfile"
	"github.com/urfave/cli/v3"
)

func main() {
	ctx := context.Background()

	cmd := &cli.Command{
		Name:                  "brick",
		Usage:                 "brick cli application",
		Description:           "brick by brick",
		Version:               "v0.0.1",
		EnableShellCompletion: true,
		Commands: []*cli.Command{
			{
				Name:   create.Name,
				Usage:  create.Usage,
				Flags:  []cli.Flag{},
				Action: create.Action,
			},
			{
				Name:   dockerfile.Name,
				Usage:  dockerfile.Usage,
				Action: dockerfile.Action,
			},
			{
				Name:   doc.Name,
				Usage:  doc.Usage,
				Action: doc.Action,
			},
		},
	}

	if err := cmd.Run(ctx, os.Args); err != nil {
		fmt.Println(err)
	}

}
