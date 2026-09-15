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

	p.setRootBounds()

	p.initBuffers()
	p.initHitMap()

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
		dirtyElements := []*Element{}
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
				dirty, quit := p.handleKeyEvent(event.KeyEvent)
				if quit {
					return nil
				}
				dirtyElements = append(dirtyElements, dirty)

			case Mouse:
				hit := p.hitMap.at(event.MouseEvent.X-1, event.MouseEvent.Y-1)
				var clicked *Element
				if event.MouseEvent.Action == MouseRelease && hit == p.mouseState.Pressed {
					clicked = hit
				}
				dirty := p.handleMouseEvent(event.MouseEvent)
				dirtyElements = append(dirtyElements, dirty...)
				if p.OnDebug != nil {
					dirtyElements = append(dirtyElements, p.OnDebug(p.debugSnapshot(event, hit, clicked))...)
				}
			}
		}
		if len(dirtyElements) > 0 {
			if err := p.paintDirtyElements(dirtyElements); err != nil {
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
	if !utf8.FullRune(buf) {
		return ParseResult{
			status: Incomplete,
		}
	}

	r, size := utf8.DecodeRune(buf)

	return ParseResult{
		event: Event{
			Kind: Key,
			KeyEvent: KeyEvent{
				Buffer:  append([]byte(nil), buf[:size]...),
				N:       size,
				KeyCode: KeyRune,
				Rune:    r,
			},
		},
		parsed: size,
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

	for i := 3; i < len(buf); i++ {
		switch {
		case buf[i] == 'M' || buf[i] == 'm':
			event, ok := parseMouseEvent(string(buf[:i+1]))
			if !ok {
				return ParseResult{status: Invalid, parsed: i + 1}
			}
			return ParseResult{event: event, status: Complete, parsed: i + 1}
		case buf[i] == ';':
			semiColonCount++
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

	switch buf[len(buf)-1] {
	case 'C':
		keyCode = KeyRight
	case 'D':
		keyCode = KeyLeft
	default:
		return ParseResult{status: Invalid, parsed: len(buf)}
	}

	return ParseResult{
		event: Event{
			Kind: Key,
			KeyEvent: KeyEvent{
				Buffer:  append([]byte{}, buf...),
				N:       len(buf),
				KeyCode: keyCode,
			},
		},
		parsed: len(buf),
		status: Complete,
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
