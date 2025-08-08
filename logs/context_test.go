package logs_test

import (
	"context"
	"testing"

	"github.com/brickingsoft/brick/logs"
	"github.com/brickingsoft/brick/pkg/mosses"
)

func TestInfo(t *testing.T) {
	log, logErr := mosses.New()
	if logErr != nil {
		t.Fatal(logErr)
	}
	defer log.Close()

	ctx := context.Background()
	ctx = logs.With(ctx, log)

	logs.Info(ctx, "hello")
}
