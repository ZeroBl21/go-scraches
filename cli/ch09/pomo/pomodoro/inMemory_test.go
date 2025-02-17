package pomodoro_test

import (
	"testing"

	"github.com/ZeroBl21/cli/ch09/pomo/pomodoro"
	"github.com/ZeroBl21/cli/ch09/pomo/pomodoro/repository"
)

func getRepo(t *testing.T) (pomodoro.Repository, func()) {
	t.Helper()

	return repository.NewInMemoryRepo(), func() {}
}
