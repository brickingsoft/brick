package mapping_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/brickingsoft/brick/pkg/mosses"
	"github.com/brickingsoft/brick/pkg/mosses/mapping"
)

func TestLog(t *testing.T) {
	logger := log.New(os.Stdout, "abc", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	logger.Println("hello world")
}

func TestLogger(t *testing.T) {
	moss, _ := mosses.New(mosses.WithLevel(mosses.DebugLevel))
	logger := mapping.Logger(context.Background(), moss)
	logger.Println("hello world")
}

func TestSetDefaultLogger(t *testing.T) {
	moss, _ := mosses.New(mosses.WithLevel(mosses.DebugLevel))
	mapping.SetDefaultLogger(context.Background(), moss)
	log.SetPrefix("prefix")
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	log.Println("hello world")
}
