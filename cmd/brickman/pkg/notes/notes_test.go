package notes_test

import (
	"context"
	"testing"

	"github.com/brickingsoft/brick/cmd/brickman/pkg/notes"
	"github.com/urfave/cli/v3"
)

func TestNew(t *testing.T) {
	logger := notes.New(&cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "log",
				Value: "info",
			},
		},
	})
	t.Log(logger != nil)
	ctx := context.Background()
	logger.Info(ctx, "1")
	notes.Warn(ctx, "2")
}
