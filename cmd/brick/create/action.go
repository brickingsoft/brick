package create

import (
	"context"

	"github.com/brickingsoft/brick/cmd/brick/pkg/log"
	"github.com/urfave/cli/v3"
)

const (
	Name  = "create"
	Usage = "create a brick project"
)

func Action(ctx context.Context, c *cli.Command) (err error) {
	loggerLevel := c.String("log")
	ctx = log.New(ctx, loggerLevel)
	log.Debug(ctx, "create ...")
	return
}
