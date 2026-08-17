package tui

import (
	"fmt"
	"strconv"
)

type ElementKind string
type Position string
type Border string

const (
	ElementBox    ElementKind = "box"
	ElementButton ElementKind = "button"
	ElementInput  ElementKind = "input"
	ElementImage  ElementKind = "image"
	ElementText   ElementKind = "text"
)

const (
	PositionAbsolute Position = "absolute"
	PositionRelative Position = "relative"
)

const (
	None Border = "none"
	Auto Border = "auto"
)

type Rect struct {
	X int
	Y int
	W int
	H int
}

type BorderChars struct {
	Top         rune
	TopLeft     rune
	TopRight    rune
	Bottom      rune
	BottomLeft  rune
	BottomRight rune
	Left        rune
	Right       rune
}

type Spacing struct {
	Top    int
	Bottom int
	Left   int
	Right  int
}

type State struct {
	Hovered  bool
	Pressed  bool
	Focused  bool
	Visible  bool
	Disabled bool
}

type Color struct {
	R, G, B uint8
	Set     bool
}

type Style struct {
	Border      Border
	BorderChars BorderChars
	Padding     Spacing
	Margin      Spacing
	Position    Position
	Fg          string
	Bg          string
	fgColor     Color
	bgColor     Color

	Hover *Style
	Press *Style
}

type EventStyle struct {
	Border      Border
	BorderChars BorderChars
	Padding     Spacing
	Margin      Spacing
	Position    Position
	Fg          string
	Bg          string
	fgColor     Color
	bgColor     Color
}

type Element struct {
	ID           string
	Rect         Rect
	ComputedRect Rect
	ContentRect  Rect
	Label        string
	Placeholder  string

	Kind     ElementKind
	State    State
	Style    Style
	Children []*Element

	OnClick      func(*Element, MouseEvent)
	OnMouseEnter func(*Element, MouseEvent)
	OnMouseOut   func(*Element, MouseEvent)
	OnInput      func(*Element, KeyEvent)
}

type TerminalDimensions struct {
	Width  int
	Height int
}

type KeyEvent struct {
	Buffer []byte
	N      int
}

type MouseEvent struct {
	Button int
	X      int
	Y      int
	Action MouseAction
}

type Event struct {
	Kind       string
	KeyEvent   KeyEvent
	MouseEvent MouseEvent
}

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

func NewBox(id string, rect Rect, style Style, children ...*Element) *Element {
	computedStyle, err := getStyle(&style, 0)
	if err != nil {
		panic(err)
	}
	return &Element{
		ID:       id,
		Kind:     ElementBox,
		Rect:     rect,
		Children: children,
		State: State{
			Visible: true,
		},
		Style: computedStyle,
	}
}

func NewButton(id string, rect Rect, label string, style Style, onClick func(*Element, MouseEvent), onMouseEnter func(*Element, MouseEvent), onMouseOut func(*Element, MouseEvent)) *Element {
	computedStyle, err := getStyle(&style, 0)
	if err != nil {
		panic(err)
	}
	return &Element{
		ID:           id,
		Kind:         ElementButton,
		Rect:         rect,
		Label:        label,
		OnClick:      onClick,
		OnMouseEnter: onMouseEnter,
		OnMouseOut:   onMouseOut,
		State: State{
			Visible: true,
		},
		Style: computedStyle,
	}
}

func NewInput(id string, rect Rect, placeholder string, style Style, onInput func(*Element, KeyEvent)) *Element {
	computedStyle, err := getStyle(&style, 0)
	if err != nil {
		panic(err)
	}
	return &Element{
		ID:          id,
		Kind:        ElementInput,
		Rect:        rect,
		Placeholder: placeholder,
		OnInput:     onInput,
		State: State{
			Visible: true,
		},
		Style: computedStyle,
	}
}

func NewText(id string, rect Rect, style Style, text string) *Element {
	computedStyle, err := getStyle(&style, 0)
	if err != nil {
		panic(err)
	}
	return &Element{
		ID:    id,
		Kind:  ElementText,
		Rect:  rect,
		Label: text,
		State: State{
			Visible: true,
		},
		Style: computedStyle,
	}
}

func NewImage(id string, rect Rect, style Style) *Element {
	computedStyle, err := getStyle(&style, 0)
	if err != nil {
		panic(err)
	}
	return &Element{
		ID:   id,
		Kind: ElementImage,
		Rect: rect,
		State: State{
			Visible: true,
		},
		Style: computedStyle,
	}
}

func AddElement(node *Element, children ...*Element) *Element {
	node.Children = append(node.Children, children...)
	return node
}

func getStyle(style *Style, depth int) (Style, error) {
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

	if depth >= 1 {
		style.Hover = nil
		style.Press = nil
		return *style, nil
	}

	if style.Hover != nil {
		hoverStyle, err := getStyle(style.Hover, 1)
		if err != nil {
			return Style{}, nil
		}
		style.Hover = &hoverStyle
	}

	if style.Press != nil {
		pressStyle, err := getStyle(style.Press, 1)
		if err != nil {
			return Style{}, nil
		}
		style.Hover = &pressStyle
	}

	return *style, nil
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
	if len(color) > 7 || color[0] != '#' {
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

func resolveStyle(el *Element) *Style {
	style := el.Style

	switch {
	case el.State.Pressed && style.Press != nil:
		style = mergeStyle(el.Style, *el.Style.Press)
	case el.State.Hovered && style.Hover != nil:
		style = mergeStyle(el.Style, *el.Style.Hover)
	}

	return &style
}

func mergeStyle(base, override Style) Style {
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
