package dockerfile

import (
	"context"

	"github.com/urfave/cli/v3"
)

const (
	Name  = "dockerfile"
	Usage = "create a dockerfile"
)

var (
	Flags = []cli.Flag{
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
