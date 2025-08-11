package main

import (
	"context"
	"fmt"
	"time"

	"github.com/brickingsoft/brick/launch/signal"
)

type Sever struct {
	ch      chan struct{}
	timeout time.Duration
}

func (s *Sever) Serve(ctx context.Context) error {
	<-s.ch
	return nil
}

func (s *Sever) Close() error {
	time.Sleep(s.timeout)
	s.ch <- struct{}{}
	return nil
}

func main() {
	srv := &Sever{
		ch:      make(chan struct{}, 1),
		timeout: 1 * time.Second,
	}

	ctx := context.Background()
	err := signal.ListenWithTimeout(ctx, srv, 5000*time.Millisecond)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("done")
}
