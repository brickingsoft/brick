package doc

import (
	"context"

	"github.com/urfave/cli/v3"
)

const (
	Name  = "doc"
	Usage = "show brick document"
)

var (
	Flags = []cli.Flag{
		&cli.IntFlag{
			Name:     "port",
			Usage:    "web server port",
			Required: false,
			Value:    9090,
		},
	}
)

func Action(ctx context.Context, c *cli.Command) (err error) {
	return
}
