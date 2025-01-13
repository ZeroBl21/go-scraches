package main

import (
	"flag"
	"fmt"
	"log"

	todo "github.com/ZeroBl21/cli/ch02"
)

const todoFileName = ".todo.json"

func main() {
	task := flag.String("task", "", "Task to be included in the To Do list")
	list := flag.Bool("list", false, "List all tasks")
	complete := flag.Int("complete", 0, "Item to be completedList all tasks")

	flag.Parse()

	l := &todo.List{}

	if err := l.Get(todoFileName); err != nil {
		log.Fatal(err)
	}

	switch {
	case *list:
		for _, item := range *l {
			if !item.Done {
				fmt.Println(item.Task)
			}
		}

	case *complete > 0:
		if err := l.Complete(*complete); err != nil {
			log.Fatal(err)
		}

		if err := l.Save(todoFileName); err != nil {
			log.Fatal(err)
		}

	case *task != "":
		l.Add(*task)

		if err := l.Save(todoFileName); err != nil {
			log.Fatal(err)
		}

	default:
		flag.PrintDefaults()
	}
}
