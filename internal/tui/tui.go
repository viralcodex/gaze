package tui

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"unicode/utf8"

	"golang.org/x/term"
)

func NewProgram(view func(TerminalDimensions) *Element, onResize func(TerminalDimensions), onUpdate func(time.Duration) bool) *Program {
	dimensions, _ := getTerminalDimensions()
	return &Program{
		hitMap:     &HitMap{},
		mouseState: MouseState{},
		postCh:     make(chan func(), 16),
		Dimensions: dimensions,
		inbuf: EventBuffer{
			state: Ground,
		},
		Canvas: Canvas{
			frame:     &Frame{},
			imageData: map[uint32]*ImageSource{},
		},
		View:     view,
		OnResize: onResize,
		OnUpdate: onUpdate,
	}
}

func (p *Program) initBuffers() {
	w, h := p.Dimensions.Width, p.Dimensions.Height
	p.frontbuff = &Buffer{Width: w, Height: h, Cells: make([]Cell, w*h)}
	p.Canvas.buff = &Buffer{Width: w, Height: h, Cells: make([]Cell, w*h)}
}

func (p *Program) initHitMap() {
	w, h := p.Dimensions.Width, p.Dimensions.Height
	p.hitMap = &HitMap{
		Width:    w,
		Height:   h,
		Cells:    make([]uint32, w*h),
		Elements: []*Element{nil},
	}
}

func (p *Program) Run() error {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}

	defer term.Restore(int(os.Stdin.Fd()), oldState)

	enterAltMode()
	defer exitAltMode()

	if err := p.updateTerminalDimensions(); err != nil {
		return err
	}

	p.Root = p.View(p.Dimensions)

	p.focused = findAutoFocus(p.Root)
	if p.focused != nil {
		p.focused.state.Focused = true
	}

	p.setRootBounds()

	p.initBuffers()
	p.initHitMap()

	hideCursorPointer()
	defer showCursorPointer()

	if err := p.render(p.Root); err != nil {
		return err
	}

	enableMouseEvents()
	defer disableMouseEvents()
	defer setCursorShape(CursorDefault)

	if err := p.eventLoop(); err != nil {
		return err
	}

	return nil
}

func findAutoFocus(el *Element) *Element {
	if el == nil {
		return nil
	}
	if el.autoFocus && (el.kind == Input || el.kind == TextArea) {
		return el
	}
	for _, child := range el.children {
		if focused := findAutoFocus(child); focused != nil {
			return focused
		}
	}
	return nil
}

func (p *Program) resize() error {
	dimensions, err := getTerminalDimensions()

	if err != nil {
		return err
	}

	p.Dimensions = dimensions

	if p.OnResize != nil {
		p.OnResize(dimensions)
	}

	p.setRootBounds()

	p.initBuffers()
	p.initHitMap()

	p.isLayoutDirty = true

	return nil
}

func (p *Program) eventLoop() error {
	resizeCh := make(chan os.Signal, 1)
	eventCh := make(chan Event, 1)

	signal.Notify(resizeCh, syscall.SIGWINCH)
	defer signal.Stop(resizeCh)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go p.readEvents(eventCh, ctx) //read the events (both key and mouse) and send through eventCh

	p.isLayoutDirty = false

	for {
		select {
		case _, ok := <-resizeCh:
			if !ok {
				return fmt.Errorf("%s", "Something wrong with resize channel")
			}

			if err := p.resize(); err != nil {
				return err
			}

		case event, ok := <-eventCh:
			if !ok {
				return fmt.Errorf("%s", "Something wrong with event channel")
			}
			switch event.Kind {
			case Key:
				quit := p.handleKeyEvent(event.KeyEvent)
				if quit {
					return nil
				}

			case Mouse:
				p.handleMouseEvent(event.MouseEvent)
			}

		case fn, ok := <-p.postCh:
			if !ok {
				return fmt.Errorf("%s", "Something wrong with post channel")
			}
			fn()
		}
		if len(p.dirtyElements) > 0 {
			if err := p.paintDirtyElements(); err != nil {
				return err
			}
		}
		if p.isLayoutDirty {
			if err := p.render(p.Root); err != nil {
				return err
			}
			p.isLayoutDirty = false
		}
	}
}

func (p *Program) readEvents(eventCh chan Event, ctx context.Context) {
	inputBuffer := make([]byte, 64)

	for {
		n, err := os.Stdin.Read(inputBuffer)

		if err != nil {
			return
		}

		events := p.inbuf.readInputBytes(inputBuffer[:n])

		for _, event := range events {
			select {
			case eventCh <- event:
			case <-ctx.Done():
				return
			}
		}
	}
}

func (inbuf *EventBuffer) readInputBytes(input []byte) []Event {
	inbuf.buffer = append(inbuf.buffer, input...)
	var events []Event

	for len(inbuf.buffer) > 0 {
		parseResult := inbuf.parseInputBytes()
		switch parseResult.status {

		case Complete: //append to events & remove parsed bytes from buffer
			parseResult.event.raw = append([]byte(nil), inbuf.buffer[:parseResult.parsed]...)
			parseResult.event.buffered = append([]byte(nil), inbuf.buffer[parseResult.parsed:]...)
			events = append(events, parseResult.event)
			inbuf.buffer = inbuf.buffer[parseResult.parsed:]

		case Invalid:
			inbuf.buffer = inbuf.buffer[max(1, parseResult.parsed):]

		case Incomplete: //new bytes still needs parsing so we return accumulated events
			if len(inbuf.buffer) >= 64 {
				inbuf.buffer = inbuf.buffer[:0]
			}
			return events
		}
	}

	return events
}

// this function checks where the stream starts and hands off parsing to a dedicated parser (key, csi, etc.)
func (inbuf *EventBuffer) parseInputBytes() ParseResult {
	switch {
	case len(inbuf.buffer) == 0:
		return ParseResult{status: Incomplete}
	case inbuf.buffer[0] != '\x1b':
		return parseKey(inbuf.buffer)
	case len(inbuf.buffer) == 1:
		return ParseResult{status: Incomplete}
	case inbuf.buffer[1] == '[':
		return parseCSI(inbuf.buffer)
	}

	return parseEscapeSequence(inbuf.buffer)
}

func parseKey(buf []byte) ParseResult {
	b := buf[0]

	var keyCode KeyCode
	var r rune
	var mods Mods

	switch {
	case b == '\r' || b == '\n':
		keyCode = KeyEnter
	case b == '\t':
		keyCode = KeyTab
	case b == 0x7F || b == 0x08:
		keyCode = KeyBackspace
	case b >= 0x01 && b <= 0x1A: // Ctrl+A..Ctrl+Z
		keyCode = KeyRune
		r = rune(b + 0x60)
		mods = Mods{Ctrl: true}
	default:
		if !utf8.FullRune(buf) {
			return ParseResult{status: Incomplete}
		}
		var size int
		r, size = utf8.DecodeRune(buf)
		return ParseResult{
			event: Event{
				Kind:     Key,
				KeyEvent: KeyEvent{KeyCode: KeyRune, Rune: r},
			},
			parsed: size,
			status: Complete,
		}
	}

	return ParseResult{
		event: Event{
			Kind:     Key,
			KeyEvent: KeyEvent{KeyCode: keyCode, Rune: r, Mods: mods},
		},
		parsed: 1,
		status: Complete,
	}
}

func parseCSI(buf []byte) ParseResult {
	if len(buf) < 3 {
		return ParseResult{
			status: Incomplete,
		}
	}
	if buf[2] == '<' {
		return parseMouseSGR(buf)
	}

	for i := 2; i < len(buf); i++ {
		if buf[i] >= 0x40 && buf[i] <= 0x7E {
			return parseCSIKey(buf[:i+1])
		}
	}

	return ParseResult{status: Incomplete}
}

func parseMouseSGR(buf []byte) ParseResult {
	semiColonCount := 0
	semiColonIndices := make([]int, 0)
	for i := 3; i < len(buf); i++ {
		switch {
		case buf[i] == 'M' || buf[i] == 'm':
			event, ok := parseMouseEvent(buf[:i+1], semiColonIndices)
			if !ok {
				return ParseResult{status: Invalid, parsed: i + 1}
			}
			return ParseResult{event: event, status: Complete, parsed: i + 1}
		case buf[i] == ';':
			semiColonCount++
			semiColonIndices = append(semiColonIndices, i)
			if semiColonCount > 2 {
				return ParseResult{status: Invalid, parsed: i + 1}
			}
		case buf[i] >= '0' && buf[i] <= '9':
		default:
			return ParseResult{
				status: Invalid,
				parsed: i + 1,
			}
		}
	}

	return ParseResult{status: Incomplete}
}

func parseEscapeSequence(buf []byte) ParseResult {
	if len(buf) < 2 {
		return ParseResult{status: Incomplete}
	}
	return ParseResult{status: Invalid, parsed: 2}
}

func parseCSIKey(buf []byte) ParseResult {
	var keyCode KeyCode
	var ok bool

	finalByte := buf[len(buf)-1] // final byte
	param1, param2 := parseCSIParams(buf[2 : len(buf)-1])

	if finalByte == '~' {
		keyCode, ok = parseTildeKey(param1)
	} else {
		keyCode, ok = parseNonTildeKey(finalByte)
	}

	if !ok {
		return ParseResult{status: Invalid, parsed: len(buf)}
	}

	return ParseResult{
		event: Event{
			Kind: Key,
			KeyEvent: KeyEvent{
				KeyCode: keyCode,
				Mods:    decodeMods(param2),
			},
		},
		parsed: len(buf),
		status: Complete,
	}
}

func parseCSIParams(buf []byte) (int, int) {
	params := []int{}
	current := 0
	for i := range buf {
		if len(params) >= 2 {
			break
		}
		if buf[i] == ';' {
			params = append(params, current)
			current = 0
		} else if buf[i] >= '0' && buf[i] <= '9' {
			current = current*10 + int(buf[i]-'0')
		}
	}
	params = append(params, current)
	param1, param2 := 0, 0
	if len(params) > 0 {
		param1 = params[0]
	}
	if len(params) > 1 {
		param2 = params[1]
	}
	return param1, param2
}

func parseTildeKey(param int) (KeyCode, bool) {
	switch param {
	case 1, 7:
		return KeyHome, true
	case 4, 8:
		return KeyEnd, true
	case 5:
		return KeyPageUp, true
	case 6:
		return KeyPageDown, true
	case 2:
		return KeyInsert, true
	case 3:
		return KeyDelete, true
	default:
		return KeyRune, false
	}
}

func parseNonTildeKey(finalByte byte) (KeyCode, bool) {
	switch finalByte {
	case 'A':
		return KeyUp, true
	case 'B':
		return KeyDown, true
	case 'C':
		return KeyRight, true
	case 'D':
		return KeyLeft, true
	case 'H':
		return KeyHome, true
	case 'F':
		return KeyEnd, true
	}
	return KeyRune, false
}

// xterm encodes the modifier param as 1 + mask (Shift=1, Alt=2, Ctrl=4).
// wire param	param-1	bits	modifiers
//
//	2			1	001		Shift
//	3			2	010		Alt
//	4			3	011		Shift+Alt
//	5			4	100		Ctrl
//	6			5	101		Ctrl+Shift
//	7			6	110		Ctrl+Alt
//	8			7	111		Ctrl+Alt+Shift
func decodeMods(param int) Mods {
	if param <= 1 {
		return Mods{}
	}

	mask := param - 1
	return Mods{
		Shift: mask&1 != 0,
		Alt:   mask&2 != 0,
		Ctrl:  mask&4 != 0,
	}
}

func (p *Program) updateTerminalDimensions() error {
	dimensions, err := getTerminalDimensions()
	if err != nil {
		return err
	}

	p.Dimensions = dimensions

	if p.OnResize != nil {
		p.OnResize(dimensions)
	}

	return nil
}

func (p *Program) setRootBounds() {
	p.Root.computedRect = Rect{
		W: p.Dimensions.Width,
		H: p.Dimensions.Height,
	}
	p.Root.contentRect = getContentRect(p.Root)
}
