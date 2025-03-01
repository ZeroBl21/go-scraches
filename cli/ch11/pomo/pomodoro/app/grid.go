package app

import (
	"github.com/mum4k/termdash/align"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/terminalapi"
)

func newGrid(
	btnSet *buttonSet,
	wid *widgets,
	sum *summary,
	term terminalapi.Terminal,
) (*container.Container, error) {
	builder := grid.New()

	// first row
	builder.Add(
		grid.RowHeightPerc(30,
			// col one
			grid.ColWidthPercWithOpts(30,
				[]container.Option{
					container.Border(linestyle.Light),
					container.BorderTitle("Press Q to Quit"),
				},
				// row 1
				grid.RowHeightPerc(80,
					grid.Widget(wid.donTimer)),
				// row 2
				grid.RowHeightPercWithOpts(20,
					[]container.Option{
						container.AlignHorizontal(align.HorizontalCenter),
					},
					grid.Widget(wid.txtTimer,
						container.AlignHorizontal(align.HorizontalCenter),
						container.AlignVertical(align.VerticalMiddle),
						container.PaddingLeftPercent(49),
					),
				),
			),
			// col two
			grid.ColWidthPerc(70,
				grid.RowHeightPerc(80,
					grid.Widget(wid.displayType, container.Border(linestyle.Light)),
				),
				grid.RowHeightPerc(20,
					grid.Widget(wid.txtInfo, container.Border(linestyle.Light)),
				),
			),
		),
	)

	// Add second row
	builder.Add(
		grid.RowHeightPerc(10,
			grid.ColWidthPerc(50,
				grid.Widget(btnSet.btnStart),
			),
			grid.ColWidthPerc(50,
				grid.Widget(btnSet.btnPause),
			),
		),
	)

	// Add third row
	builder.Add(
		grid.RowHeightPerc(60,
			grid.ColWidthPerc(30,
				grid.Widget(sum.bcDay,
					container.Border(linestyle.Light),
					container.BorderTitle("Daily Summary (minutes)"),
				),
			),
			grid.ColWidthPerc(70,
				grid.Widget(sum.lcWeekly,
					container.Border(linestyle.Light),
					container.BorderTitle("Weekly Summary"),
				),
			),
		),
	)

	gridOpts, err := builder.Build()
	if err != nil {
		return nil, err
	}

	c, err := container.New(term, gridOpts...)
	if err != nil {
		return nil, err
	}

	return c, nil
}
