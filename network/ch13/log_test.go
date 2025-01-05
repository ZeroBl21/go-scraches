package ch13

import (
	"bytes"
	"fmt"
	"log"
	"os"
)

func Example_log() {
	l := log.New(os.Stdout, "example: ", log.Lshortfile)
	l.Print("logging to standard output")

	// Output:
	// example: log_test.go:12: logging to standard output
}

func Example_logMultiWriter() {
	logFile := new(bytes.Buffer)
	w := NewSustainedMultiWriter(os.Stdout, logFile)
	l := log.New(w, "example: ", log.Lshortfile|log.Lmsgprefix)

	fmt.Println("standard output:")
	l.Print("Canada is south of Detroit")

	fmt.Print("\nlog file contents:\n", logFile.String())

	// Output:
	// standard output:
	// log_test.go:24: example: Canada is south of Detroit
	//
	// log file contents:
	// log_test.go:24: example: Canada is south of Detroit
}
