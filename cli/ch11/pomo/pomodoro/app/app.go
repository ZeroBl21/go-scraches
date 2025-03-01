package app

import (
	"context"
	"image"
	"time"

	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/terminal/tcell"
	"github.com/mum4k/termdash/terminal/terminalapi"

	"github.com/ZeroBl21/cli/ch10/pomo/pomodoro"
)

type App struct {
	ctx        context.Context
	controller *termdash.Controller
	term       *tcell.Terminal
	size       image.Point

	redrawCh chan bool
	errCh    chan error
}

func New(config *pomodoro.IntervalConfig) (*App, error) {
	ctx, cancel := context.WithCancel(context.Background())

	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == 'q' || k.Key == 'Q' {
			cancel()
		}
	}

	redrawCh := make(chan bool)
	errCh := make(chan error)

	wid, err := newWidgets(ctx, errCh)
	if err != nil {
		return nil, err
	}

	sum, err := newSummary(ctx, config, errCh)
	if err != nil {
		return nil, err
	}

	btnSet, err := newButtonSet(ctx, config, wid, sum, redrawCh, errCh)
	if err != nil {
		return nil, err
	}

	term, err := tcell.New()
	if err != nil {
		return nil, err
	}

	container, err := newGrid(btnSet, wid, sum, term)
	if err != nil {
		return nil, err
	}

	controller, err := termdash.NewController(
		term,
		container,
		termdash.KeyboardSubscriber(quitter),
	)
	if err != nil {
		return nil, err
	}

	return &App{
		ctx:        ctx,
		controller: controller,
		term:       term,

		redrawCh: redrawCh,
		errCh:    errCh,
	}, nil
}

func (a *App) Run() error {
	defer a.term.Close()
	defer a.controller.Close()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.redrawCh:
			if err := a.controller.Redraw(); err != nil {
				return err
			}

		case <-ticker.C:
			if err := a.resize(); err != nil {
				return err
			}

		case err := <-a.errCh:
			if err != nil {
				return err
			}

		case <-a.ctx.Done():
			return nil
		}
	}
}

func (a *App) resize() error {
	if a.size.Eq(a.term.Size()) {
		return nil
	}

	a.size = a.term.Size()
	if err := a.term.Clear(); err != nil {
		return err
	}

	return a.controller.Redraw()
}
