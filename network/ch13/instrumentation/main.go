package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/ZeroBl21/network/ch13/instrumentation/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	metricsAddr = flag.String("metrics", "127.0.0.1:8081", "metrics listen address")
	webAddr     = flag.String("web", "127.0.0.1:8082", "web listen address")
)

func main() {
	flag.Parse()

	mux := http.NewServeMux()

	mux.Handle("/metrics/", promhttp.Handler())

	if err := newHTTPServer(*metricsAddr, mux, nil); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Metrics listening on %q ... \n", *metricsAddr)

	err := newHTTPServer(*webAddr, http.HandlerFunc(helloHandler), connStateMetrics)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Web listening on %q ...\n\n", *webAddr)

	clients := 500
	gets := 100
	wg := new(sync.WaitGroup)

	fmt.Printf("Spawning %d connections to make %d request each ...", clients, gets)
	for i := 0; i < clients; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			c := &http.Client{
				// Don't reuse cache connections
				Transport: http.DefaultTransport.(*http.Transport).Clone(),
			}

			for j := 0; j < gets; j++ {
				resp, err := c.Get(fmt.Sprintf("http://%s/", *webAddr))
				if err != nil {
					log.Fatal(err)
				}

				io.Copy(io.Discard, resp.Body)
				resp.Body.Close()
			}
		}()
	}

	wg.Wait()

	fmt.Print(" done.\n\n")

	resp, err := http.Get(fmt.Sprintf("http://%s/metrics", *metricsAddr))
	if err != nil {
		log.Fatal(err)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()

	metricsPrefix := fmt.Sprintf("%s_%s", *metrics.Namespace, *metrics.Subsystem)
	fmt.Println("Current Metrics:")

	for _, line := range bytes.Split(b, []byte("\n")) {
		if bytes.HasPrefix(line, []byte(metricsPrefix)) {
			fmt.Printf("%s\n", line)
		}
	}
}

func helloHandler(w http.ResponseWriter, _ *http.Request) {
	metrics.Requests.Add(1)
	defer func(start time.Time) {
		metrics.RequestDuration.Observe(time.Since(start).Seconds())
	}(time.Now())

	if _, err := w.Write([]byte("Hello")); err != nil {
		metrics.WriteErrors.Add(1)
	}
}

func newHTTPServer(
	addr string,
	mux http.Handler,
	stateFunc func(net.Conn, http.ConnState),
) error {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		IdleTimeout:       time.Minute,
		ReadHeaderTimeout: 30 * time.Second,
		ConnState:         stateFunc,
	}

	go func() { log.Fatal(srv.Serve(l)) }()

	return nil
}

func connStateMetrics(_ net.Conn, state http.ConnState) {
	switch state {
	case http.StateNew:
		metrics.OpenConnections.Add(1)
	case http.StateClosed:
		metrics.OpenConnections.Add(-1)
	}
}
