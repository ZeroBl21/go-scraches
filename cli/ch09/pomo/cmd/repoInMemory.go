package cmd

import (
	"github.com/ZeroBl21/cli/ch09/pomo/pomodoro"
	"github.com/ZeroBl21/cli/ch09/pomo/pomodoro/repository"
)

func getRepo() (pomodoro.Repository, error) {
	return repository.NewInMemoryRepo(), nil
}
