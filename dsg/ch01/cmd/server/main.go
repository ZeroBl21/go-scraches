package main

import (
	"log"

	"github.com/ZeroBl21/dsg/ch01/proglog/internal/server"
)

func main() {
	srv := server.NewHTTPServer(":8080")
	log.Fatal(srv.ListenAndServe())
}
