package tui

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

type Style struct {
	Border      Border
	BorderChars BorderChars
	Padding     Spacing
	Margin      Spacing
	Foreground  string
	Background  string
	Position    Position
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

	OnClick func(*Element, MouseEvent)
	OnHover func(*Element)
	OnInput func(*Element, KeyEvent)
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
	Press  bool
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
	return &Element{
		ID:       id,
		Kind:     ElementBox,
		Rect:     rect,
		Children: children,
		State: State{
			Visible: true,
		},
		Style: getStyle(&style),
	}
}

func NewButton(id string, rect Rect, label string, style Style, onClick func(*Element, MouseEvent)) *Element {
	return &Element{
		ID:      id,
		Kind:    ElementButton,
		Rect:    rect,
		Label:   label,
		OnClick: onClick,
		State: State{
			Visible: true,
		},
		Style: getStyle(&style),
	}
}

func NewInput(id string, rect Rect, placeholder string, style Style, onInput func(*Element, KeyEvent)) *Element {
	return &Element{
		ID:          id,
		Kind:        ElementInput,
		Rect:        rect,
		Placeholder: placeholder,
		OnInput:     onInput,
		State: State{
			Visible: true,
		},
		Style: getStyle(&style),
	}
}

func NewText(id string, rect Rect, text string) *Element {
	return &Element{
		ID:    id,
		Kind:  ElementText,
		Rect:  rect,
		Label: text,
		State: State{
			Visible: true,
		},
	}
}

func NewImage(id string, rect Rect, style Style) *Element {
	return &Element{
		ID:   id,
		Kind: ElementImage,
		Rect: rect,
		State: State{
			Visible: true,
		},
		Style: getStyle(&style),
	}
}

func AddElement(node *Element, children ...*Element) *Element {
	node.Children = append(node.Children, children...)
	return node
}

func getStyle(style *Style) Style {
	if style.Border == Auto {
		setBorderChars(style)
	} else {
		style.BorderChars = BorderChars{}
	}
	return *style
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
