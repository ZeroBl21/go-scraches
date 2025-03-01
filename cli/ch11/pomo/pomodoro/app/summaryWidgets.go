package app

import (
	"context"
	"math"
	"time"

	"github.com/ZeroBl21/cli/ch10/pomo/pomodoro"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/barchart"
	"github.com/mum4k/termdash/widgets/linechart"
)

type summary struct {
	bcDay    *barchart.BarChart
	lcWeekly *linechart.LineChart

	updateDaily  chan bool
	updateWeekly chan bool
}

func newSummary(
	ctx context.Context,
	config *pomodoro.IntervalConfig,
	errorCh chan<- error,
) (*summary, error) {
	s := &summary{
		updateDaily:  make(chan bool),
		updateWeekly: make(chan bool),
	}

	var err error

	s.bcDay, err = newBarChart(ctx, config, s.updateDaily, errorCh)
	if err != nil {
		return nil, err
	}

	s.lcWeekly, err = newLineChart(ctx, config, s.updateWeekly, errorCh)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func newBarChart(
	ctx context.Context,
	config *pomodoro.IntervalConfig,
	updateCh <-chan bool,
	errCh chan<- error,
) (*barchart.BarChart, error) {
	bc, err := barchart.New(
		barchart.ShowValues(),
		barchart.BarColors([]cell.Color{
			cell.ColorPurple,
			cell.ColorGreen,
		}),
		barchart.ValueColors([]cell.Color{
			cell.ColorBlack,
			cell.ColorBlack,
		}),
		barchart.Labels([]string{
			"Pomodoro",
			"Break",
		}),
	)
	if err != nil {
		return nil, err
	}

	updateWidget := func() error {
		ds, err := pomodoro.DailySummary(time.Now(), config)
		if err != nil {
			return err
		}

		return bc.Values([]int{
			int(ds[0].Minutes()),
			int(ds[1].Minutes()),
		},
			int(math.Max(
				ds[0].Minutes(),
				ds[1].Minutes())*1.1)+1,
		)
	}

	go func() {
		for {
			select {
			case <-updateCh:
				errCh <- updateWidget()
			case <-ctx.Done():
				return
			}
		}
	}()

	if err != updateWidget() {
		return nil, err
	}

	return bc, nil
}

func newLineChart(
	ctx context.Context,
	config *pomodoro.IntervalConfig,
	updateCh <-chan bool,
	errCh chan<- error,
) (*linechart.LineChart, error) {
	lc, err := linechart.New(
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorPurple)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorCyan)),
		linechart.YAxisFormattedValues(
			linechart.ValueFormatterSingleUnitDuration(time.Second, 0),
		),
	)
	if err != nil {
		return nil, err
	}

	updateWidget := func() error {
		ws, err := pomodoro.RangeSummary(time.Now(), 7, config)
		if err != nil {
			return err
		}

		err = lc.Series(ws[0].Name, ws[0].Values,
			linechart.SeriesCellOpts(cell.FgColor(cell.ColorBlue)),
			linechart.SeriesXLabels(ws[1].Labels),
		)
		if err != nil {
			return err
		}

		return lc.Series(ws[1].Name, ws[1].Values,
			linechart.SeriesCellOpts(cell.FgColor(cell.ColorYellow)),
			linechart.SeriesXLabels(ws[1].Labels),
		)
	}

	go func() {
		for {
			select {
			case <-updateCh:
				errCh <- updateWidget()
			case <-ctx.Done():
				return
			}
		}
	}()

	if err := updateWidget(); err != nil {
		return nil, err
	}

	return lc, nil
}

func (s *summary) update(redrawCh chan<- bool) {
	s.updateDaily <- true
	s.updateWeekly <- true

	redrawCh <- true
}
