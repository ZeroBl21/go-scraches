//go:build inmemory
// +build inmemory

package pomodoro_test

import (
	"testing"

	"github.com/ZeroBl21/cli/ch10/pomo/pomodoro"
	"github.com/ZeroBl21/cli/ch10/pomo/pomodoro/repository"
)

func getRepo(t *testing.T) (pomodoro.Repository, func()) {
	t.Helper()

	return repository.NewInMemoryRepo(), func() {}
}
