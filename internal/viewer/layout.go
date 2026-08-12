package viewer

import (
	"fmt"
	"gaze/internal/tui"
)

var layoutStyle = tui.Style{
	Position: tui.PositionRelative,
}

var buttonStyle = tui.Style{
	Position: tui.PositionRelative,
	Border:   tui.Auto,
	BorderChars: tui.BorderChars{
		Top:         '━',
		TopLeft:     '┏',
		TopRight:    '┓',
		Bottom:      '━',
		BottomLeft:  '┗',
		BottomRight: '┛',
		Left:        '┃',
		Right:       '┃',
	},
	Padding: tui.Spacing{Left: 2, Right: 2, Top: 1, Bottom: 1},
}

func createLayout() {
	buttonGroup := terminalState.Dimensions.Width - terminalState.Dimensions.Width % 4
	buttonWidth := buttonGroup / 4

	terminalState.Root = tui.AddElement(terminalState.Root,
		tui.NewBox("buttonGroup", tui.Rect{
			W: buttonGroup,
		}, layoutStyle,
			tui.NewButton("zoom+", tui.Rect{
				W: buttonWidth,
			}, "+",
				buttonStyle,
				func(el *tui.Element, event tui.MouseEvent) {
					fmt.Println("clicked", el.ID)
				},
			),
			tui.NewButton("zoom-", tui.Rect{
				X: buttonWidth,
				W: buttonWidth,
			}, "-", buttonStyle, func(el *tui.Element, e tui.MouseEvent) {
				fmt.Println("clicked", el.ID)
			},
			),
			tui.NewButton("rotate+", tui.Rect{
				X: 2 * buttonWidth,
				W: buttonWidth,
			}, "-90deg", buttonStyle, func(el *tui.Element, e tui.MouseEvent) {
				fmt.Println("clicked", el.ID)
			},
			),
			tui.NewButton("rotate-", tui.Rect{
				X: 3 * buttonWidth,
				W: buttonWidth,
			}, "+90deg", buttonStyle, func(el *tui.Element, e tui.MouseEvent) {
				fmt.Println("clicked", el.ID)
			},
			),
		),
		tui.NewImage("image", tui.Rect{
			X: (terminalState.Dimensions.Width - img.Rect.Cols) / 2,
			Y: (terminalState.Dimensions.Height + 10 - img.Rect.Rows) / 2,
			W: terminalState.Dimensions.Width,
			H: terminalState.Dimensions.Height - 10,
		}, layoutStyle,
		),
	)
}

// TODO: change this impl later
func updateLayout() {
	updateTerminalDimensions()
	terminalState.Root = tui.NewBox("root", tui.Rect{X: 1, Y: 1, W: terminalState.Dimensions.Width, H: terminalState.Dimensions.Height}, layoutStyle)
	getImageRect()
	createLayout()
}

func updateTerminalDimensions() {
	dimensions, err := tui.GetTerminalDimensions()

	if err != nil {
		return
	}

	terminalState.Dimensions = dimensions
}
