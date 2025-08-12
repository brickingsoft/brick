package create

import (
	"context"

	"github.com/urfave/cli/v3"
)

const (
	Name  = "create"
	Usage = "create a brick project"
)

var (
	Flags = []cli.Flag{
		&cli.StringFlag{
			Name:     "mod",
			Usage:    "project module path",
			Required: true,
		},
		&cli.StringFlag{
			Name:     "out",
			Usage:    "output dir path",
			Required: false,
		},
	}
)

func Action(ctx context.Context, c *cli.Command) (err error) {
	return
}
