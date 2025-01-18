package todo

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

type item struct {
	Task        string
	Done        bool
	CreatedAt   time.Time
	CompletedAt time.Time
}

type List []item

func (l *List) Add(task string) {
	t := item{
		Task:        task,
		Done:        false,
		CreatedAt:   time.Now(),
		CompletedAt: time.Time{},
	}

	*l = append(*l, t)
}

func (l *List) Complete(idx int) error {
	ls := *l

	if idx <= 0 || idx > len(ls) {
		return fmt.Errorf("Item %d does not exists", idx)
	}

	ls[idx-1].Done = true
	ls[idx-1].CompletedAt = time.Now()

	return nil
}

func (l *List) Delete(idx int) error {
	ls := *l
	if idx <= 0 || idx > len(ls) {
		return fmt.Errorf("Item %d does not exists", idx)
	}

	*l = append(ls[:idx-1], ls[idx:]...)

	return nil
}

func (l *List) Save(filename string) error {
	js, err := json.Marshal(l)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, js, 0644)
}

func (l *List) Get(filename string) error {
	file, err := os.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return err
	}

	if len(file) == 0 {
		return nil
	}

	return json.Unmarshal(file, l)
}

func (l *List) String() string {
	formatted := ""

	for idx, task := range *l {
		prefix := "[ ] "
		if task.Done {
			prefix = "[X] "
		}

		formatted += fmt.Sprintf("%s%d: %s\n", prefix, idx+1, task.Task)
	}

	return formatted
}

func (l *List) Pending() string {
	formatted := ""

	for idx, task := range *l {
		if !task.CompletedAt.IsZero() {
			continue
		}

		prefix := "[ ] "
		if task.Done {
			prefix = "[X] "
		}

		formatted += fmt.Sprintf("%s%d: %s\n", prefix, idx+1, task.Task)
	}

	return formatted
}
