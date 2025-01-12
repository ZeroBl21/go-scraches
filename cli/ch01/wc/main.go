package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
)

const (
	CountLines = 1 << iota
	CountBytes
)

func main() {
	lines := flag.Bool("l", false, "Count lines")
	bytes := flag.Bool("b", false, "Bytes lines")

	flag.Parse()

	flags := 0
	if *lines {
		flags |= CountLines
	}
	if *bytes {
		flags |= CountBytes
	}

	fmt.Println(count(os.Stdin, flags))
}

func count(r io.Reader, flags int) int {
	scanner := bufio.NewScanner(r)

	if flags&CountLines != 0 {
		scanner.Split(bufio.ScanLines)
	} else if flags&CountBytes != 0 {
		scanner.Split(bufio.ScanBytes)
	} else {
		scanner.Split(bufio.ScanWords)
	}

	wc := 0

	for scanner.Scan() {
		wc++
	}

	return wc
}
