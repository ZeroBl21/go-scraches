package app

import (
	"context"
	"fmt"

	"github.com/ZeroBl21/cli/ch09/pomo/pomodoro"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/button"
)

type buttonSet struct {
	btnStart *button.Button
	btnPause *button.Button
}

func newButtonSet(
	ctx context.Context,
	config *pomodoro.IntervalConfig,
	w *widgets,
	redrawCh chan<- bool,
	errCh chan<- error,
) (*buttonSet, error) {
	startInterval := func() {
		i, err := pomodoro.GetInterval(config)
		errCh <- err

		start := func(i pomodoro.Interval) {
			msg := "Take a break"
			if i.Category == pomodoro.CategoryPomodoro {
				msg = "Focus on your task"
			}

			w.update([]int{}, i.Category, msg, "", redrawCh)
		}

		periodic := func(i pomodoro.Interval) {
			w.update(
				[]int{int(i.ActualDuration), int(i.PlannedDuration)},
				"", "", fmt.Sprint(i.PlannedDuration-i.ActualDuration), redrawCh)
		}

		end := func(i pomodoro.Interval) {
			w.update([]int{}, "", "Nothing running...", "", redrawCh)
		}

		errCh <- i.Start(ctx, config, start, periodic, end)
	}

	pauseInterval := func() {
		i, err := pomodoro.GetInterval(config)
		if err != nil {
			errCh <- err
			return
		}

		if err := i.Pause(config); err != nil {
			if err == pomodoro.ErrIntervalNotRunning {
				return
			}
			errCh <- err
			return
		}

		w.update([]int{}, "", "Paused... press start to continue", "", redrawCh)
	}

	btnStart, err := button.New("(S)tart", func() error {
		go startInterval()
		return nil
	},
		button.GlobalKey('s'),
		button.WidthFor("(P)ause"),
		button.Height(2),
	)
	if err != nil {
		return nil, err
	}

	btnPause, err := button.New("(P)ause", func() error {
		go pauseInterval()
		return nil
	},
		button.FillColor(cell.ColorNumber(220)),
		button.GlobalKey('p'),
		button.Height(2),
	)
	if err != nil {
		return nil, err
	}

	return &buttonSet{btnStart, btnPause}, nil
}
