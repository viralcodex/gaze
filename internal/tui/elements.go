package tui

type ElementKind string

const (
	ElementBox    ElementKind = "box"
	ElementButton ElementKind = "button"
	ElementInput  ElementKind = "input"
	ElementImage  ElementKind = "image"
	ElementText   ElementKind = "text"
)

type Rect struct {
	X int
	Y int
	W int
	H int
}

type ElementState struct {
	Hovered  bool
	Pressed  bool
	Focused  bool
	Visible  bool
	Disabled bool
}

type ElementStyle struct {
	Border     bool
	Padding    int
	Foreground string
	Background string
}

type Element struct {
	ID          string
	Rect        Rect
	Label       string
	Placeholder string

	Kind     ElementKind
	State    ElementState
	Style    ElementStyle
	Children []*Element

	OnClick func(*Element)
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

func NewBox(id string, rect Rect, children ...*Element) *Element {
	return &Element{
		ID:       id,
		Kind:     ElementBox,
		Rect:     rect,
		Children: children,
		State: ElementState{
			Visible: true,
		},
	}
}

func NewButton(id string, rect Rect, label string, onClick func(*Element)) *Element {
	return &Element{
		ID:      id,
		Kind:    ElementButton,
		Rect:    rect,
		Label:   label,
		OnClick: onClick,
		State: ElementState{
			Visible: true,
		},
	}
}

func NewInput(id string, rect Rect, placeholder string, onInput func(*Element, KeyEvent)) *Element {
	return &Element{
		ID:          id,
		Kind:        ElementInput,
		Rect:        rect,
		Placeholder: placeholder,
		OnInput:     onInput,
		State: ElementState{
			Visible: true,
		},
	}
}

func NewText(id string, rect Rect, text string) *Element {
	return &Element{
		ID:    id,
		Kind:  ElementText,
		Rect:  rect,
		Label: text,
		State: ElementState{
			Visible: true,
		},
	}
}

func NewImage(id string, rect Rect) *Element {
	return &Element{
		ID:   id,
		Kind: ElementImage,
		Rect: rect,
		State: ElementState{
			Visible: true,
		},
	}
}

func AddElement(node *Element, children ...*Element) *Element {
	node.Children = append(node.Children, children...)
	return node
}
