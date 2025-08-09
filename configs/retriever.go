package configs

import (
	"context"

	"github.com/brickingsoft/brick/pkg/mists"
)

type Retriever interface {
	Retrieve(ctx context.Context, params map[string]string) (mist mists.Mist, err error)
}

var (
	retrievers = make(map[string]Retriever)
)

func RegisterRetriever(name string, retriever Retriever) {
	if retriever == nil {
		return
	}
	if name == "" {
		return
	}
	retrievers[name] = retriever
}

func GetRetriever(name string) (v Retriever, has bool) {
	retriever, ok := retrievers[name]
	if !ok {
		return
	}
	has = retriever != nil
	if has {
		v = retriever
	}
	return
}

type MultiLevelRetriever struct{}

func (retriever *MultiLevelRetriever) Retrieve(ctx context.Context, params map[string]string) (mist mists.Mist, err error) {
	// params.config = dir
	// params.active = dev
	return
}
