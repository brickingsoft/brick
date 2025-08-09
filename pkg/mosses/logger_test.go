package mosses_test

import (
	"context"
	"testing"
	"time"

	"github.com/brickingsoft/brick/pkg/mosses"
)

func TestNew(t *testing.T) {
	log, logErr := mosses.New()
	if logErr != nil {
		t.Fatal(logErr)
	}
	defer log.Close()
	log = log.Group("g1").Attr(mosses.String("hello", "world"))
	log = log.Group("g2").Attr(mosses.String("hello", "world")).Attr(mosses.String("hello", "mosses"), mosses.Int("int", 1))
	ctx := context.Background()
	log.Info(ctx, "hello")
	log.Info(ctx, "world")
}

func TestNewAsyncHandler(t *testing.T) {
	log, logErr := mosses.New(mosses.WithHandler(mosses.NewAsyncHandler(mosses.NewStandardOutColorfulHandler(), mosses.AsyncHandlerOptions{
		Workers:          4,
		WorkerChanBuffer: 0,
		CloseTimeout:     1 * time.Second,
	})))
	if logErr != nil {
		t.Fatal(logErr)
	}
	defer log.Close()
	log = log.Group("test").Attr(mosses.String("hello", "world"))
	ctx := context.Background()
	log.Info(ctx, "hello")

}

func TestNewStandardOutJson(t *testing.T) {
	log, logErr := mosses.New(mosses.WithHandler(mosses.NewStandardOutJsonHandler()))
	if logErr != nil {
		t.Fatal(logErr)
	}
	defer log.Close()
	log = log.Group("test").Attr(mosses.String("hello", "world"))
	ctx := context.Background()
	log.Info(ctx, "hello")
	log.Info(ctx, "world")
}

func TestNewStandardOutText(t *testing.T) {
	log, logErr := mosses.New(mosses.WithHandler(mosses.NewStandardOutHandler()))
	if logErr != nil {
		t.Fatal(logErr)
	}
	defer log.Close()
	log = log.Group("test").Attr(mosses.String("hello", "world"))
	ctx := context.Background()
	log.Info(ctx, "hello")
	log.Info(ctx, "world")
}

func TestMoss_CallerSkipShift(t *testing.T) {
	log, logErr := mosses.New()
	if logErr != nil {
		t.Fatal(logErr)
	}
	defer log.Close()
	log = log.Group("test").Attr(mosses.String("hello", "world"))
	ctx := context.Background()

	var skip = func(ctx context.Context, logger mosses.Logger) {
		logger.Info(ctx, "hello")
	}

	log = log.CallerSkipShift(1)
	skip(ctx, log)
}
