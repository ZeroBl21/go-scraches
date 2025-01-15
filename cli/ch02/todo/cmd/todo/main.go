package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	todo "github.com/ZeroBl21/cli/ch02"
)

var todoFileName = ".todo.json"

func main() {
	add := flag.Bool("add", false, "Add task to the to do list")
	list := flag.Bool("list", false, "List all tasks")
	complete := flag.Int("complete", 0, "Item to be completed")

	flag.Parse()

	if os.Getenv("TODO_FILENAME") != "" {
		todoFileName = os.Getenv("TODO_FILENAME")
	}

	l := &todo.List{}

	if err := l.Get(todoFileName); err != nil {
		log.Fatal(err)
	}

	switch {
	case *add:
		t, err := getTask(os.Stdin, flag.Args()...)
		if err != nil {
			log.Fatal(err)
		}
		l.Add(t)

		if err := l.Save(todoFileName); err != nil {
			log.Fatal(err)
		}

	case *list:
		fmt.Print(l)

	case *complete > 0:
		if err := l.Complete(*complete); err != nil {
			log.Fatal(err)
		}

		if err := l.Save(todoFileName); err != nil {
			log.Fatal(err)
		}

	default:
		flag.PrintDefaults()
	}
}

func getTask(r io.Reader, args ...string) (string, error) {
	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}

	s := bufio.NewScanner(r)
	s.Scan()

	if err := s.Err(); err != nil {
		return "", err
	}

	if len(s.Text()) == 0 {
		return "", fmt.Errorf("Task cannot be blank")
	}

	return s.Text(), nil
}
