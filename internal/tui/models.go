package tui

import (
	"bytes"
	"time"
)

type Program struct {
	Root *Element

	hitMap        *HitMap
	mouseState    MouseState
	isLayoutDirty bool
	isPaintDirty  bool
	inbuf         EventBuffer
	frontbuff     *Buffer
	focused       *Element
	cursor        Cursor

	Canvas     Canvas
	Dimensions TerminalDimensions

	View     func(TerminalDimensions) *Element
	OnResize func(TerminalDimensions)
	OnUpdate func(time.Duration) bool
	OnDebug  func(DebugSnapshot) []*Element
}

// this is where elements draw on
type Canvas struct {
	buff      *Buffer
	frame     *Frame
	imageData map[uint32]*ImageSource
}

//---element structs---

type Element struct {
	ID          string
	Rect        Rect
	Label       string
	Placeholder string
	MaxLength   int
	Style       Style
	Layout      Layout
	ImageRef    ImageRef //set by the client, tui doesn't create this

	measuredRect Rect
	computedRect Rect
	contentRect  Rect
	kind         ElementKind
	state        state
	children     []*Element

	inputState InputState

	Draw         func(*Element, *Canvas) error
	OnClick      func(*Element, MouseEvent)
	OnMouseEnter func(*Element, MouseEvent)
	OnMouseOut   func(*Element, MouseEvent)
	OnInput      func(*Element, KeyEvent)
}

// ---props structs---
// Zero values mean defaults: Rect{} auto-sizes, empty Style/Layout use
// existing defaults, and nil handlers mean no callback.

type BoxProps struct {
	ID     string
	Rect   Rect
	Style  Style
	Layout Layout
}

type ButtonProps struct {
	ID     string
	Rect   Rect
	Label  string
	Style  Style
	Layout Layout

	OnClick      func(*Element, MouseEvent)
	OnMouseEnter func(*Element, MouseEvent)
	OnMouseOut   func(*Element, MouseEvent)
}

type InputProps struct {
	ID          string
	Rect        Rect
	Placeholder string
	MaxLength   int // 0 means unlimited
	Style       Style
	Layout      Layout

	OnInput func(*Element, KeyEvent)
}

type TextProps struct {
	ID     string
	Rect   Rect
	Text   string
	Style  Style
	Layout Layout
}

type ImageProps struct {
	ID       string
	Rect     Rect
	Style    Style
	Layout   Layout
	ImageRef ImageRef
}

type ElementKind string
type Border string
type Position uint8
type Cursor uint8

const (
	CursorDefault Cursor = iota
	CursorPointer
	CursorText
)

const (
	Box    ElementKind = "box"
	Button ElementKind = "button"
	Input  ElementKind = "input"
	Image  ElementKind = "image"
	Text   ElementKind = "text"
)

const (
	Relative Position = iota
	Absolute
)

const (
	None Border = "none"
	Auto Border = "auto"
)

type EventKind string

const (
	Mouse EventKind = "mouse"
	Key   EventKind = "key"
)

type Rect struct {
	X int
	Y int
	W int
	H int
}

type Style struct {
	Border      Border
	BorderChars BorderChars
	Padding     Spacing
	Margin      Spacing
	Position    Position
	Cursor      Cursor
	Fg          string
	Bg          string
	fgColor     Color
	bgColor     Color

	Hover *StateStyle
	Press *StateStyle
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

// ---Layout structs---
type Direction uint8

const (
	Row Direction = iota
	Column
)

type Justify uint8

const (
	Start Justify = iota
	Center
	End
	SpaceBetween
	SpaceEvenly
	SpaceAround
)

type Align uint8

const (
	AlignStart Align = iota
	AlignCenter
	AlignEnd
	AlignStretch
)

type Layout struct {
	Direction Direction
	Justify   Justify
	Align     Align
	Gap       int
	Grow      int
}

type DirectionVals struct {
	main  []int
	cross []int
}

// ---state structs---
type state struct {
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

type StateStyle struct {
	Fg string
	Bg string

	fgColor Color
	bgColor Color
}

type TerminalDimensions struct {
	Width  int
	Height int
}

type InputState struct {
	Text         []rune
	Cursor       int
	Row          int
	Col          int
	ScrollOffset int
}

type KeyCode uint16

const (
	KeyRune KeyCode = iota
	KeyLeft
	KeyRight
	KeyUp
	KeyDown
	KeyEnter
	KeyBackspace
	KeyDelete
	KeyHome
	KeyEnd
	KeyTab
	KeyEscape
)

type KeyEvent struct {
	Buffer  []byte //will remove later (buf, n) if not needed
	N       int
	KeyCode KeyCode
	Rune    rune
}

type MouseEvent struct {
	Button int
	X      int
	Y      int
	Action MouseAction
}

type Event struct {
	Kind       EventKind
	KeyEvent   KeyEvent
	MouseEvent MouseEvent
	raw        []byte
	buffered   []byte
}

type DebugSnapshot struct {
	Parsed   string
	Buffered string
	Mouse    MouseEvent
	Hit      string
	Hovered  string
	Pressed  string
	Clicked  string
	Focused  string
}

type EventBuffer struct {
	buffer []byte
	state  ParseState
}

type ParseState uint8

const (
	Ground ParseState = iota
	Escape
	Csi
	Sgr
)

type ParserStatus uint8

const (
	Incomplete ParserStatus = iota
	Complete
	Invalid
)

type ParseResult struct {
	event  Event
	parsed int
	status ParserStatus
}

// ---events structs---
const (
	MouseMove MouseAction = iota
	MousePress
	MouseRelease
)

type HitMap struct {
	Width    int
	Height   int
	Cells    []uint32
	Elements []*Element
}

type MouseState struct {
	Hovered *Element
	Pressed *Element
}

type MouseAction uint8

// ---image---
type ImageRef struct {
	ID         uint32
	Dimensions ImageDimensions
	Position   ObjectPosition
}

// ObjectPosition aligns the fitted image within the element's resolved box,
// per axis. Zero value is AlignStart/AlignStart (top-left).
type ObjectPosition struct {
	X Align
	Y Align
}

type ImageSource struct {
	ID          uint32
	Data        []byte
	NeedsUpload bool
}

type ImageDimensions struct {
	Width  int
	Height int
}

// ---misc---
type Frame struct {
	wbuf bytes.Buffer
}

type Cell struct {
	Rune rune
	Fg   Color
	Bg   Color
	/// more fields (bold, underline, italics)...
}

type Buffer struct {
	Width, Height int
	Cells         []Cell
}
