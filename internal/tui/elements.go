package tui

import (
	"fmt"
	"strconv"
)

var defaultBorders = BorderChars{
	Top:         '─',
	TopLeft:     '┌',
	TopRight:    '┐',
	Bottom:      '─',
	BottomLeft:  '└',
	BottomRight: '┘',
	Left:        '│',
	Right:       '│',
}

// newElement builds the common runtime node shared by every constructor:
func newElement(kind ElementKind, id string, rect Rect, style Style, layout Layout, focusable bool) *Element {
	computedStyle, err := getStyle(&style)
	if err != nil {
		panic(err)
	}
	return &Element{
		ID:        id,
		kind:      kind,
		Rect:      rect,
		Style:     computedStyle,
		Layout:    layout,
		state:     state{Visible: true},
		Focusable: focusable,
	}
}

func NewBox(props BoxProps, children ...*Element) *Element {
	box := newElement(Box, props.ID, props.Rect, props.Style, props.Layout, props.Focusable)
	box.children = children
	box.Draw = drawBox
	for _, child := range children {
		child.Parent = box
	}
	return box
}

func NewButton(props ButtonProps) *Element {
	el := newElement(Button, props.ID, props.Rect, props.Style, props.Layout, props.Focusable)
	el.Label = props.Label
	el.OnClick = props.OnClick
	el.OnMouseEnter = props.OnMouseEnter
	el.OnMouseOut = props.OnMouseOut
	el.Draw = drawButton
	return el
}

func NewInput(props InputProps) *Element {
	el := newElement(Input, props.ID, props.Rect, props.Style, props.Layout, props.Focusable)
	el.Placeholder = props.Placeholder
	el.MaxLength = props.MaxLength
	el.autoFocus = props.AutoFocus
	el.inputState = InputState{Text: [][]rune{{}}}
	el.OnInput = props.OnInput
	el.Draw = drawInput
	return el
}

func NewTextArea(props InputProps) *Element {
	el := newElement(TextArea, props.ID, props.Rect, props.Style, props.Layout, props.Focusable)
	el.Placeholder = props.Placeholder
	el.MaxLength = props.MaxLength
	el.autoFocus = props.AutoFocus
	el.inputState = InputState{Line: 0, Text: [][]rune{{}}, DesiredX: -1}
	el.OnInput = props.OnInput
	el.Draw = drawTextArea
	return el
}

func NewText(props TextProps) *Element {
	el := newElement(Text, props.ID, props.Rect, props.Style, props.Layout, props.Focusable)
	el.Label = props.Text
	el.Draw = drawText
	return el
}

func NewImage(props ImageProps) *Element {
	el := newElement(Image, props.ID, props.Rect, props.Style, props.Layout, props.Focusable)
	el.ImageRef = props.ImageRef
	el.Draw = drawImage
	return el
}

func AddElement(node *Element, children ...*Element) *Element {
	for _, child := range children {
		child.Parent = node
	}
	node.children = append(node.children, children...)
	return node
}

func getStyle(style *Style) (Style, error) {
	if style.Border == Auto {
		setBorderChars(style)
	} else {
		style.BorderChars = BorderChars{}
	}

	fgColor, err := parseHexColor(style.Fg)
	if err != nil {
		return Style{}, err
	}

	bgColor, err := parseHexColor(style.Bg)
	if err != nil {
		return Style{}, err
	}

	style.fgColor = fgColor
	style.bgColor = bgColor

	if style.Hover != nil {
		err := getStateStyle(style.Hover)
		if err != nil {
			return Style{}, err
		}
	}

	if style.Press != nil {
		err := getStateStyle(style.Press)
		if err != nil {
			return Style{}, err
		}
	}

	if style.Select != nil {
		err := getStateStyle(style.Select)
		if err != nil {
			return Style{}, err
		}
	}

	return *style, nil
}

func getStateStyle(style *StateStyle) error {
	if style == nil {
		return nil
	}

	var err error

	style.fgColor, err = parseHexColor(style.Fg)
	if err != nil {
		return err
	}

	style.bgColor, err = parseHexColor(style.Bg)
	return err
}

func mergeStyle(base Style, override StateStyle) Style {
	if override.Fg != "" {
		base.Fg = override.Fg
		base.fgColor = override.fgColor
	}

	if override.Bg != "" {
		base.Bg = override.Bg
		base.bgColor = override.bgColor
	}

	return base
}

func setBorderChars(style *Style) {
	if style.BorderChars.Top == 0 {
		style.BorderChars.Top = defaultBorders.Top
	}
	if style.BorderChars.TopLeft == 0 {
		style.BorderChars.TopLeft = defaultBorders.TopLeft
	}
	if style.BorderChars.TopRight == 0 {
		style.BorderChars.TopRight = defaultBorders.TopRight
	}
	if style.BorderChars.Bottom == 0 {
		style.BorderChars.Bottom = defaultBorders.Bottom
	}
	if style.BorderChars.BottomLeft == 0 {
		style.BorderChars.BottomLeft = defaultBorders.BottomLeft
	}
	if style.BorderChars.BottomRight == 0 {
		style.BorderChars.BottomRight = defaultBorders.BottomRight
	}
	if style.BorderChars.Left == 0 {
		style.BorderChars.Left = defaultBorders.Left
	}
	if style.BorderChars.Right == 0 {
		style.BorderChars.Right = defaultBorders.Right
	}
}

func parseHexColor(color string) (Color, error) {
	if color == "" {
		return Color{}, nil
	}
	if len(color) != 7 || color[0] != '#' {
		return Color{}, fmt.Errorf("Invalid Hex Color format")
	}

	raw, err := strconv.ParseUint(color[1:], 16, 24)

	if err != nil {
		return Color{}, fmt.Errorf("invalid color %q: %w", color, err)
	}

	return Color{
		R:   uint8(raw >> 16),
		G:   uint8(raw >> 8),
		B:   uint8(raw),
		Set: true,
	}, nil
}

// styleFor overlays the element's active state (press/hover/select) onto its
// cached, inheritance-resolved base style. It walks no ancestors, so partial
// repaints can call it as cheaply as a full draw.
func styleFor(el *Element) Style {
	style := el.resolvedStyle

	switch {
	case el.state.Pressed && style.Press != nil:
		style = mergeStyle(style, *style.Press)
	case el.state.Hovered && style.Hover != nil:
		style = mergeStyle(style, *style.Hover)
	case el.state.Selected && style.Select != nil:
		style = mergeStyle(style, *style.Select)
	}

	return style
}
