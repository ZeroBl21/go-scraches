package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	todo "github.com/ZeroBl21/cli/ch02"
)

const todoFileName = ".todo.json"

func main() {
	l := &todo.List{}

	if err := l.Get(todoFileName); err != nil {
		log.Fatal(err)
	}

	switch {
	case len(os.Args) == 1:
		for _, item := range *l {
			fmt.Println(item.Task)
		}
	default:
		item := strings.Join(os.Args[1:], " ")

		l.Add(item)

		if err := l.Save(todoFileName); err != nil {
			log.Fatal(err)
		}
	}
}
