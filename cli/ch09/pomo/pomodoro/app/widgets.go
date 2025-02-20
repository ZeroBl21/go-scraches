package app

import (
	"context"
	"strings"

	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/widgets/donut"
	"github.com/mum4k/termdash/widgets/segmentdisplay"
	"github.com/mum4k/termdash/widgets/text"
)

type widgets struct {
	donTimer    *donut.Donut
	displayType *segmentdisplay.SegmentDisplay
	txtInfo     *text.Text
	txtTimer    *text.Text

	updateDonTimer chan []int
	updateTxtInfo  chan string
	updateTxtTimer chan string
	updateTxtType  chan string
}

func newWidgets(ctx context.Context, errorCh chan<- error) (*widgets, error) {
	w := &widgets{
		updateDonTimer: make(chan []int),
		updateTxtInfo:  make(chan string),
		updateTxtTimer: make(chan string),
		updateTxtType:  make(chan string),
	}

	var err error

	w.donTimer, err = newDonut(ctx, w.updateDonTimer, errorCh)
	if err != nil {
		return nil, err
	}

	w.displayType, err = newSegmentDisplay(ctx, w.updateTxtType, errorCh)
	if err != nil {
		return nil, err
	}

	w.txtInfo, err = newText(ctx, w.updateTxtInfo, errorCh)
	if err != nil {
		return nil, err
	}

	w.txtTimer, err = newText(ctx, w.updateTxtTimer, errorCh)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (w *widgets) update(
	timer []int,
	txtType, txtInfo, txtTimer string,
	redrawCh chan<- bool,
) {
	if txtInfo != "" {
		w.updateTxtInfo <- txtInfo
	}

	if txtType != "" {
		w.updateTxtType <- txtType
	}

	if txtTimer != "" {
		w.updateTxtTimer <- txtTimer
	}

	if len(timer) > 0 {
		w.updateDonTimer <- timer
	}

	redrawCh <- true
}

func newText(
	ctx context.Context,
	updateText <-chan string,
	errorCh chan<- error,
) (*text.Text, error) {
	txt, err := text.New()
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case t := <-updateText:
				txt.Reset()
				errorCh <- txt.Write(t)

			case <-ctx.Done():
				return
			}
		}
	}()

	return txt, nil
}

func newDonut(
	ctx context.Context,
	donUpdater <-chan []int,
	errCh chan<- error,
) (*donut.Donut, error) {
	don, err := donut.New(
		donut.Clockwise(),
		donut.CellOpts(cell.FgColor(cell.ColorPurple)),
	)
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case d := <-donUpdater:
				if d[0] <= d[1] {
					errCh <- don.Absolute(d[0], d[1])
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	return don, nil
}

func newSegmentDisplay(
	ctx context.Context,
	updateText <-chan string,
	errCh chan<- error,
) (*segmentdisplay.SegmentDisplay, error) {
	sd, err := segmentdisplay.New()
	if err != nil {
		return nil, err
	}

	go func() {
		for {
			select {
			case t := <-updateText:
				if t == "" {
					t = " "
				}

				errCh <- sd.Write([]*segmentdisplay.TextChunk{
					segmentdisplay.NewChunk(strings.ToUpper(t)),
				})

			case <-ctx.Done():
				return
			}
		}
	}()

	return sd, nil
}
