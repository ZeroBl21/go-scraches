package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/ZeroBl21/network/ch12/housework/v1"
	"google.golang.org/grpc"
)

var addr, certFn, keyFn string

func init() {
	flag.StringVar(&addr, "address", ":34443", "listen address")
	flag.StringVar(&certFn, "cert", "cert.pem", "certificate file")
	flag.StringVar(&keyFn, "key", "key.pem", "private key file")
}

func main() {
	flag.Parse()

	server := grpc.NewServer()
	rosie := &Rosie{
		mu:                           sync.Mutex{},
		chores:                       []*housework.Chore{},
		UnimplementedRobotMaidServer: housework.UnimplementedRobotMaidServer{},
	}
	housework.RegisterRobotMaidServer(server, rosie)

	cert, err := tls.LoadX509KeyPair(certFn, keyFn)
	if err != nil {
		log.Fatal(err)
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Listening for TLS connection on %s ...", addr)

	err = server.Serve(
		tls.NewListener(
			listener,
			&tls.Config{
				Certificates:             []tls.Certificate{cert},
				CurvePreferences:         []tls.CurveID{tls.CurveP256},
				MinVersion:               tls.VersionTLS12,
				PreferServerCipherSuites: true,
				NextProtos:               []string{"h2"},
			},
		),
	)
	if err != nil {
		log.Fatal(err)
	}
}
