package main

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"
)

func TestDialContextCancelFanOut(t *testing.T) {
	ctx, cancel := context.WithDeadline(
		context.Background(),
		time.Now().Add(10*time.Second),
	)

	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal()
	}
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if nil == err {
			conn.Close()
		}
	}()

	dial := func(
		ctx context.Context,
		address string,
		response chan int,
		id int,
		wg *sync.WaitGroup,
	) {
		defer wg.Done()

		var dialer net.Dialer

		conn, err := dialer.DialContext(ctx, "tcp", address)
		if err != nil {
			return
		}
		conn.Close()

		select {
		case <-ctx.Done():
		case response <- id:
		}
	}

	res := make(chan int)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go dial(ctx, listener.Addr().String(), res, i+1, &wg)
	}

	response := <-res
	cancel()
	wg.Wait()
	close(res)

	if ctx.Err() != context.Canceled {
		t.Errorf("expected canceled context; actual: %s", ctx.Err())
	}

	t.Logf("dialer %d retrieved the resource", response)
}
