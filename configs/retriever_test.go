package configs_test

import (
	"context"
	"embed"
	"os"
	"testing"

	"github.com/brickingsoft/brick/configs"
)

var (
	//go:embed testcases/*.yaml
	testcases embed.FS
)

func TestNewFileMultiLevelRetriever(t *testing.T) {
	os.Args = []string{os.Args[0]}
	t.Log("os.Args", os.Args)
	r := configs.NewFileMultiLevelRetriever(configs.WithRetrieverDir("./testcases"))
	testFileMultiLevelRetrieverRetrieve(t, r)

	r = configs.NewFileMultiLevelRetriever(configs.WithRetrieverEmbedDir(&testcases))
	testFileMultiLevelRetrieverRetrieve(t, r)
}

func testFileMultiLevelRetrieverRetrieve(t *testing.T, r configs.Retriever) {
	ctx := context.Background()
	config, configErr := r.Retrieve(ctx)
	if configErr != nil {
		t.Fatal(configErr)
	}
	t.Log(config)

	envErr := os.Setenv("BRICK_ACTIVE", "dev")
	if envErr != nil {
		t.Fatal(envErr)
	}
	config, configErr = r.Retrieve(ctx)
	if configErr != nil {
		t.Fatal(configErr)
	}
	t.Log(config)

	envErr = os.Setenv("BRICK_ACTIVE", "")
	if envErr != nil {
		t.Fatal(envErr)
	}
	os.Args = []string{os.Args[0], "-active=prod"}
	config, configErr = r.Retrieve(ctx)
	if configErr != nil {
		t.Fatal(configErr)
	}
	t.Log(config)
}
