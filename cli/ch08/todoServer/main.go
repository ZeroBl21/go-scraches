package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"
)

func greet(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World! %s", time.Now())
}

func main() {
	host := flag.String("h", "localhost", "Server host")
	port := flag.Int("p", 8080, "Server port")
	todoFile := flag.String("f", "todoServer.json", "todo JSON file")

	flag.Parse()

	server := &http.Server{
		Addr:         net.JoinHostPort(*host, strconv.Itoa(*port)),
		Handler:      newMux(*todoFile),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
