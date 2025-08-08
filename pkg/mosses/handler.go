package mosses

import (
	"context"
	"io"
	"math/bits"
	"os"
	"sync"
	"time"

	"github.com/mattn/go-colorable"
)

type Handler interface {
	Handle(context.Context, *Record)
	Close() error
}

func NewStandardOutHandler() Handler {
	return &StandardOutHandler{
		writer:  os.Stdout,
		locker:  &sync.Mutex{},
		encoder: NewTextRecordEncoder(),
	}
}

func NewStandardOutColorfulHandler() Handler {
	w := colorable.NewColorableStdout()
	return &StandardOutHandler{
		writer:  w,
		locker:  &sync.Mutex{},
		encoder: NewColorfulTextRecordEncoder(),
	}
}

func NewStandardOutJsonHandler() Handler {
	return &StandardOutHandler{
		writer:  os.Stdout,
		locker:  &sync.Mutex{},
		encoder: NewJsonRecordEncoder(),
	}
}

type StandardOutHandler struct {
	encoder RecordEncoder
	locker  sync.Locker
	writer  io.Writer
}

func (handler *StandardOutHandler) Handle(_ context.Context, record *Record) {
	if record == nil {
		return
	}
	b := handler.encoder.Encode(record)
	b = append(b, '\n')
	handler.locker.Lock()
	_, _ = handler.writer.Write(b)
	handler.locker.Unlock()
}

func (handler *StandardOutHandler) Close() error {
	return nil
}

type AsyncHandlerOptions struct {
	Workers          int           `json:"workers" yaml:"workers"`
	WorkerChanBuffer int           `json:"workerChanBuffer" yaml:"workerChanBuffer"`
	CloseTimeout     time.Duration `json:"closeTimeout" yaml:"closeTimeout"`
}

type asyncHandlerRequest struct {
	ctx    context.Context
	record *Record
}

func floorPow2(n uint32) uint32 {
	if n == 0 {
		return 0
	}
	shift := bits.Len(uint(n)) - 1
	return 1 << shift
}

func NewAsyncHandler(proxy Handler, options AsyncHandlerOptions) Handler {
	workers := options.Workers
	workers = int(floorPow2(uint32(workers)))
	if workers < 1 {
		workers = 1
	}

	workerChanBuffer := options.WorkerChanBuffer
	if workerChanBuffer < 1 {
		workerChanBuffer = 1024 * 8
	}
	requestChs := make([]chan asyncHandlerRequest, workers)
	for i := 0; i < workers; i++ {
		requestChs[i] = make(chan asyncHandlerRequest, workerChanBuffer)
	}
	closeTimeout := options.CloseTimeout
	if closeTimeout < 1 {
		closeTimeout = 0
	}

	handler := &AsyncHandler{
		proxy:          proxy,
		requestChs:     requestChs,
		requestChsIdx:  0,
		requestChsSize: uint64(workers),
		locker:         new(sync.Mutex),
		closed:         false,
		closeTimeout:   closeTimeout,
		closeCh:        make(chan struct{}, workers),
	}

	for i := 0; i < workers; i++ {
		go func(proxy Handler, ch <-chan asyncHandlerRequest, cch chan<- struct{}) {
			for {
				req, ok := <-ch
				if !ok {
					cch <- struct{}{}
					break
				}
				proxy.Handle(req.ctx, req.record)
			}
		}(proxy, handler.requestChs[i], handler.closeCh)
	}

	return handler
}

type AsyncHandler struct {
	proxy          Handler
	requestChs     []chan asyncHandlerRequest
	requestChsIdx  uint64
	requestChsSize uint64
	locker         sync.Locker
	closed         bool
	closeTimeout   time.Duration
	closeCh        chan struct{}
}

func (handler *AsyncHandler) Handle(ctx context.Context, record *Record) {
	if record == nil {
		return
	}
	handler.locker.Lock()
	if handler.closed {
		handler.locker.Unlock()
		return
	}
	i := handler.requestChsIdx & handler.requestChsSize
	handler.requestChsIdx++
	ch := handler.requestChs[i]
	ch <- asyncHandlerRequest{ctx: ctx, record: record}
	handler.locker.Unlock()
}

func (handler *AsyncHandler) Close() error {
	handler.locker.Lock()
	if handler.closed {
		handler.locker.Unlock()
		return nil
	}
	handler.closed = true
	for _, ch := range handler.requestChs {
		close(ch)
	}
	handler.locker.Unlock()

	if handler.closeTimeout < 1 {
		for i := 0; i < len(handler.requestChs); i++ {
			<-handler.closeCh
		}
	} else {
		done := false
		timer := time.NewTimer(handler.closeTimeout)
		for i := 0; i < len(handler.requestChs); i++ {
			select {
			case <-handler.closeCh:
				break
			case <-timer.C:
				done = true
				break
			}
			if done {
				break
			}
		}
		timer.Stop()
	}
	return nil
}
