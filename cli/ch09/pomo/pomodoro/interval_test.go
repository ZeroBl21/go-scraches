package pomodoro_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ZeroBl21/cli/ch09/pomo/pomodoro"
)

func TestNewConfig(t *testing.T) {
	testCases := []struct {
		name     string
		input    [3]time.Duration
		expected pomodoro.IntervalConfig
	}{
		{
			name: "Default",
			expected: pomodoro.IntervalConfig{
				PomodoroDuration:   25 * time.Minute,
				ShortBreakDuration: 5 * time.Minute,
				LongBreakDuration:  15 * time.Minute,
			},
		},
		{
			name: "SingleInput",
			input: [3]time.Duration{
				20 * time.Minute,
			},
			expected: pomodoro.IntervalConfig{
				PomodoroDuration:   20 * time.Minute,
				ShortBreakDuration: 5 * time.Minute,
				LongBreakDuration:  15 * time.Minute,
			},
		},
		{
			name: "MultiInput",
			input: [3]time.Duration{
				20 * time.Minute,
				10 * time.Minute,
				12 * time.Minute,
			},
			expected: pomodoro.IntervalConfig{
				PomodoroDuration:   20 * time.Minute,
				ShortBreakDuration: 10 * time.Minute,
				LongBreakDuration:  12 * time.Minute,
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			var repo pomodoro.Repository
			config := pomodoro.NewConfig(
				repo,
				tt.input[0],
				tt.input[1],
				tt.input[2],
			)

			if config.PomodoroDuration != tt.expected.PomodoroDuration {
				t.Errorf("Expected Pomodoro Duration %q, got %q instead\n",
					tt.expected.PomodoroDuration, config.PomodoroDuration)
			}
			if config.ShortBreakDuration != tt.expected.ShortBreakDuration {
				t.Errorf("Expected Short Break Duration %q, got %q instead\n",
					tt.expected.ShortBreakDuration, config.ShortBreakDuration)
			}
			if config.LongBreakDuration != tt.expected.LongBreakDuration {
				t.Errorf("Expected Long Break Duration %q, got %q instead\n",
					tt.expected.LongBreakDuration, config.LongBreakDuration)
			}
		})
	}
}

func TestGetInterval(t *testing.T) {
	repo, cleanup := getRepo(t)
	defer cleanup()

	const duration = 1 * time.Millisecond
	config := pomodoro.NewConfig(repo, 3*duration, duration, 2*duration)

	for i := 1; i <= 16; i++ {
		var expCategory string
		var expDuration time.Duration

		switch {
		case i%2 != 0:
			expCategory = pomodoro.CategoryPomodoro
			expDuration = 3 * duration

		case i%8 == 0:
			expCategory = pomodoro.CategoryLongBreak
			expDuration = 2 * duration

		case i%2 == 0:
			expCategory = pomodoro.CategoryShortBreak
			expDuration = duration
		}

		testName := fmt.Sprintf("%s%d", expCategory, i)
		t.Run(testName, func(t *testing.T) {
			res, err := pomodoro.GetInterval(config)
			if err != nil {
				t.Errorf("Expected no error, got %q.\n", err)
			}

			noop := func(pomodoro.Interval) {}

			err = res.Start(context.Background(), config, noop, noop, noop)
			if err != nil {
				t.Fatal(err)
			}

			if res.Category != expCategory {
				t.Errorf("Expected category %q, got %q.\n",
					expCategory, res.Category)
			}

			if res.PlannedDuration != expDuration {
				t.Errorf("Expected duration %q, got %q.\n",
					expDuration, res.PlannedDuration)
			}

			if res.State != pomodoro.StateNotStarted {
				t.Errorf("Expected state %d, got %d instead.\n",
					pomodoro.StateNotStarted, res.State)
			}

			ui, err := repo.ByID(res.ID)
			if err != nil {
				t.Errorf("Expected no error. Got %q.\n", err)
			}

			if ui.State != pomodoro.StateDone {
				t.Errorf("Expected state %d, got %d instead.\n",
					pomodoro.StateDone, res.State)
			}
		})
	}
}

func TestPause(t *testing.T) {
	const duration = 2 * time.Second

	repo, cleanup := getRepo(t)
	defer cleanup()

	config := pomodoro.NewConfig(repo, duration, duration, duration)

	testCases := []struct {
		name        string
		start       bool
		expState    int
		expDuration time.Duration
	}{
		{
			name: "NotStarted", start: false,
			expState: pomodoro.StateNotStarted, expDuration: 0,
		},
		{
			name: "Paused", start: true,
			expState: pomodoro.StatePaused, expDuration: duration / 2,
		},
	}
	expErr := pomodoro.ErrIntervalNotRunning

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			i, err := pomodoro.GetInterval(config)
			if err != nil {
				t.Fatal(err)
			}

			start := func(pomodoro.Interval) {}
			periodic := func(i pomodoro.Interval) {
				if err := i.Pause(config); err != nil {
					t.Fatal(err)
				}
			}
			end := func(pomodoro.Interval) {
				t.Error("End callback should not be executed")
			}

			if tt.start {
				if err := i.Start(ctx, config, start, periodic, end); err != nil {
					t.Fatal(err)
				}
			}

			i, err = pomodoro.GetInterval(config)
			if err != nil {
				t.Fatal(err)
			}

			err = i.Pause(config)
			if err != nil {
				if !errors.Is(err, expErr) {
					t.Fatalf("Expected error %q, got %q instead.", expErr, err)
				}
			}

			if err == nil {
				t.Errorf("Expected error %q got nil", expErr)
			}

			i, err = repo.ByID(i.ID)
			if err != nil {
				t.Fatal(err)
			}

			if i.State != tt.expState {
				t.Errorf("Expected state %d, got %d instead.\n",
					tt.expState, i.State)
			}

			if i.ActualDuration != tt.expDuration {
				t.Errorf("Expected state %d, got %d instead.\n",
					tt.expDuration, i.ActualDuration)
			}

			cancel()
		})
	}
}

func TestStart(t *testing.T) {
	const duration = 2 * time.Second

	repo, cleanup := getRepo(t)
	defer cleanup()

	config := pomodoro.NewConfig(repo, duration, duration, duration)

	testCases := []struct {
		name        string
		cancel      bool
		expState    int
		expDuration time.Duration
	}{
		{
			name: "Finish", cancel: false,
			expState: pomodoro.StateDone, expDuration: duration,
		},
		{
			name: "Cancel", cancel: true,
			expState: pomodoro.StateCancelled, expDuration: duration / 2,
		},
	}

	for _, tt := range testCases {
		ctx, cancel := context.WithCancel(context.Background())

		i, err := pomodoro.GetInterval(config)
		if err != nil {
			t.Fatal(err)
		}

		start := func(i pomodoro.Interval) {
			if i.State != pomodoro.StateRunning {
				t.Errorf("Expected state %d, got %d instead.\n",
					pomodoro.StateRunning, i.State)
			}

			if i.ActualDuration >= i.PlannedDuration {
				t.Errorf("Expected ActualDuration %q. less than planned %q.\n",
					i.ActualDuration, i.PlannedDuration)
			}
		}

		periodic := func(i pomodoro.Interval) {
			if i.State != pomodoro.StateRunning {
				t.Errorf("Expected state %d, got %d instead.\n",
					pomodoro.StateRunning, i.State)
			}
			if tt.cancel {
				cancel()
			}
		}

		end := func(i pomodoro.Interval) {
			if i.State != tt.expState {
				t.Errorf("Expected state %d, got %d instead.\n",
					tt.expState, i.State)
			}
			if tt.cancel {
				t.Errorf("End callback should not be executed when canceled")
			}
		}

		if err := i.Start(ctx, config, start, periodic, end); err != nil {
			t.Fatal(err)
		}

		i, err = repo.ByID(i.ID)
		if err != nil {
			t.Fatal(err)
		}

		if i.State != tt.expState {
			t.Errorf("Expected state %d, got %d instead.\n",
				tt.expState, i.State)
		}

		if i.ActualDuration != tt.expDuration {
			t.Errorf("Expected state %d, got %d instead.\n",
				tt.expState, i.State)
		}

		cancel()
	}
}
