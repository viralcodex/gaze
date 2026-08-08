package viewer

import (
	"fmt"
	"gaze/internal/tui"
	"math"
)

func createLayout() {
	terminalState.Root = tui.AddElement(terminalState.Root,
		tui.NewBox("buttons", tui.Rect{
			X: 2,
			Y: 2,
			W: terminalState.Dimensions.Width - 2,
			H: 4,
		},
			tui.NewButton("zoom+", tui.Rect{
				X: 2,
				Y: 2,
				W: int(math.Floor(float64(terminalState.Dimensions.Width)/3) - 1),
				H: 3,
			}, "+", func(el *tui.Element) {
				fmt.Println("clicked", el.ID)
			},
			),
			tui.NewButton("zoom-", tui.Rect{
				X: int(math.Floor(float64(terminalState.Dimensions.Width)/3)) + 1,
				Y: 2,
				W: int(math.Floor(float64(terminalState.Dimensions.Width) / 3)),
				H: 3,
			}, "-", func(el *tui.Element) {
				fmt.Println("clicked", el.ID)
			},
			),
			tui.NewButton("rotate", tui.Rect{
				X: int(2*math.Floor(float64(terminalState.Dimensions.Width)/3)) + 1,
				Y: 2,
				W: int(math.Floor(float64(terminalState.Dimensions.Width) / 3)),
				H: 3,
			}, "+90deg", func(el *tui.Element) {
				fmt.Println("clicked", el.ID)
			},
			),
		),
		tui.NewImage("image", tui.Rect{
			X: (terminalState.Dimensions.Width-img.Rect.Cols)/2 + 1,
			Y: 6,
			W: terminalState.Dimensions.Width,
			H: terminalState.Dimensions.Height - 6,
		},
		),
	)
}

// TODO: change this impl later
func updateLayout() {
	updateTerminalDimensions()
	terminalState.Root = tui.NewBox("root", tui.Rect{X: 1, Y: 1, W: terminalState.Dimensions.Width, H: terminalState.Dimensions.Height}, nil)
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
