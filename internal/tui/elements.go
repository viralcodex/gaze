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

func NewBox(props BoxProps, children ...*Element) *Element {
	computedStyle, err := getStyle(&props.Style)
	if err != nil {
		panic(err)
	}
	return &Element{
		ID:       props.ID,
		kind:     Box,
		Rect:     props.Rect,
		children: children,
		state: state{
			Visible: true,
		},
		Style:  computedStyle,
		Layout: props.Layout,
		Draw:   drawBox,
	}
}

func NewButton(props ButtonProps) *Element {
	computedStyle, err := getStyle(&props.Style)
	if err != nil {
		panic(err)
	}
	return &Element{
		ID:           props.ID,
		kind:         Button,
		Rect:         props.Rect,
		Label:        props.Label,
		OnClick:      props.OnClick,
		OnMouseEnter: props.OnMouseEnter,
		OnMouseOut:   props.OnMouseOut,
		state: state{
			Visible: true,
		},
		Style:  computedStyle,
		Layout: props.Layout,
		Draw:   drawButton,
	}
}

func NewInput(props InputProps) *Element {
	computedStyle, err := getStyle(&props.Style)
	if err != nil {
		panic(err)
	}
	return &Element{
		ID:          props.ID,
		kind:        Input,
		Rect:        props.Rect,
		Placeholder: props.Placeholder,
		MaxLength:   props.MaxLength,
		state: state{
			Visible: true,
		},
		inputState: InputState{
			Text:   []rune{},
			Cursor: 0,
		},
		Style:   computedStyle,
		Layout:  props.Layout,
		Draw:    drawInput,
		OnInput: props.OnInput,
	}
}

func NewText(props TextProps) *Element {
	computedStyle, err := getStyle(&props.Style)
	if err != nil {
		panic(err)
	}
	return &Element{
		ID:    props.ID,
		kind:  Text,
		Rect:  props.Rect,
		Label: props.Text,
		state: state{
			Visible: true,
		},
		Style:  computedStyle,
		Layout: props.Layout,
		Draw:   drawText,
	}
}

func NewImage(props ImageProps) *Element {
	computedStyle, err := getStyle(&props.Style)
	if err != nil {
		panic(err)
	}
	return &Element{
		ID:       props.ID,
		kind:     Image,
		ImageRef: props.ImageRef,
		Rect:     props.Rect,
		state: state{
			Visible: true,
		},
		Style:  computedStyle,
		Layout: props.Layout,
		Draw:   drawImage,
	}
}

func AddElement(node *Element, children ...*Element) *Element {
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

// resolves style at runtime
func resolveStyle(el *Element) *Style {
	style := el.Style

	switch {
	case el.state.Pressed && style.Press != nil:
		style = mergeStyle(el.Style, *el.Style.Press)
	case el.state.Hovered && style.Hover != nil:
		style = mergeStyle(el.Style, *el.Style.Hover)
	}

	if el.Style.Position != Absolute && el.Style.Position != Relative {
		el.Style.Position = Relative
	}

	return &style
}
