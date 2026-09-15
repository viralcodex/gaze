package viewer

import (
	"gaze/internal/tui"
)

var layoutStyle = tui.Style{
	// Position: tui.Absolute,
	Bg: "#1E2428",
	Fg: "#E8ECEF",
}

var buttonStyle = tui.Style{
	Border: tui.Auto,
	// Position: tui.Relative,
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
	Padding: tui.Spacing{Left: 0, Right: 0, Top: 1, Bottom: 1},
	Cursor:  tui.CursorPointer,
	Bg:      "#28434A",
	Fg:      "#FFD166",
	Hover: &tui.StateStyle{
		Bg: "#31535B",
		Fg: "#F4F7F8",
	},

	Press: &tui.StateStyle{
		Bg: "#FFD166",
		Fg: "#1E2428",
	},
}

var debugParsed *tui.Element
var debugBuffered *tui.Element
var debugMouse *tui.Element

func createLayout() *tui.Element {
	buttonWidth := terminalState.Dimensions.Width / 5

	root := tui.NewBox(tui.BoxProps{
		ID:    "root",
		Style: layoutStyle,
		Layout: tui.Layout{
			Direction: tui.Column,
			Align:     tui.AlignStretch,
		},
	},
		tui.NewBox(tui.BoxProps{
			ID:   "buttonGroup",
			Rect: tui.Rect{W: terminalState.Dimensions.Width},
			Style: tui.Style{
				Bg:      "#e2822f",
				Padding: tui.Spacing{Top: 1, Bottom: 1},
			},
			Layout: tui.Layout{
				Direction: tui.Row,
				Justify:   tui.SpaceEvenly,
			},
		},
			tui.NewButton(tui.ButtonProps{
				ID:    "zoom+",
				Rect:  tui.Rect{W: buttonWidth},
				Label: "button1",
				Style: buttonStyle,
				OnClick: func(el *tui.Element, event tui.MouseEvent) {
					zoomImage(el)
				},
			}),
			tui.NewButton(tui.ButtonProps{
				ID:    "zoom-",
				Rect:  tui.Rect{W: buttonWidth},
				Label: "button2",
				Style: buttonStyle,
				OnClick: func(el *tui.Element, e tui.MouseEvent) {
					zoomImage(el)
				},
			}),
			tui.NewButton(tui.ButtonProps{
				ID:    "rotate+",
				Rect:  tui.Rect{W: buttonWidth},
				Label: "button3",
				Style: buttonStyle,
				OnClick: func(el *tui.Element, e tui.MouseEvent) {
					rotateImage(el)
				},
			}),
			tui.NewButton(tui.ButtonProps{
				ID:    "rotate-",
				Rect:  tui.Rect{W: buttonWidth},
				Label: "button4",
				Style: buttonStyle,
				OnClick: func(el *tui.Element, e tui.MouseEvent) {
					rotateImage(el)
				},
			}),
		),
		tui.NewImage(tui.ImageProps{
			ID: "image",
			Style: tui.Style{
				Bg: "#8fff96",
				Padding: tui.Spacing{
					Top:    1,
					Bottom: 1,
					Right:  1,
					Left:   1,
				},
			},
			Layout: tui.Layout{
				Grow: 1,
			},
			ImageRef: tui.ImageRef{
				ID:         nextImageID,
				Dimensions: tui.ImageDimensions(terminalState.ImageData.Dimensions),
				Position: tui.ObjectPosition{
					X: tui.AlignCenter,
					Y: tui.AlignCenter,
				},
			},
		}),
		tui.NewInput(tui.InputProps{
			ID:          "input",
			Rect:        tui.Rect{W: terminalState.Dimensions.Width},
			Placeholder: "enter text here",
			MaxLength:   1000,
			Style: tui.Style{
				Bg:     "#3c9dff",
				Fg:     "#000000",
				Border: tui.Auto,
				Cursor: tui.CursorText,
				Padding: tui.Spacing{
					Top:    0,
					Bottom: 0,
					Left:   1,
					Right:  1,
				},
			},
		}),
		tui.NewInput(tui.InputProps{
			ID:          "input2",
			Rect:        tui.Rect{W: terminalState.Dimensions.Width},
			Placeholder: "enter text here",
			MaxLength:   1000,
			Style: tui.Style{
				Bg:     "#39e12a",
				Fg:     "#a20416",
				Border: tui.Auto,
				Cursor: tui.CursorText,
				Padding: tui.Spacing{
					Top:    0,
					Bottom: 0,
					Left:   1,
					Right:  1,
				},
			},
		}),
		tui.NewBox(tui.BoxProps{
			ID:   "debug",
			Rect: tui.Rect{H: 5},
			Style: tui.Style{
				Border:  tui.Auto,
				Padding: tui.Spacing{Left: 1, Right: 1},
				Bg:      "#171B1D",
				Fg:      "#A9BAC1",
			},
			Layout: tui.Layout{Direction: tui.Column, Align: tui.AlignStretch},
		},
			newDebugText("debugParsed"),
			newDebugText("debugBuffered"),
			newDebugText("debugMouse"),
		),
	)

	terminalState.Root = root
	return terminalState.Root
}

func newDebugText(id string) *tui.Element {
	text := tui.NewText(tui.TextProps{
		ID:    id,
		Rect:  tui.Rect{H: 1},
		Style: tui.Style{Fg: "#A9BAC1", Bg: "#171B1D"},
	})
	switch id {
	case "debugParsed":
		debugParsed = text
	case "debugBuffered":
		debugBuffered = text
	case "debugMouse":
		debugMouse = text
	}
	return text
}
