package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

type TerminalState struct {
	event      Event
	text       []byte
	dimensions TerminalDimensions
}

type TerminalDimensions struct {
	width  int
	height int
}

type KeyEvent struct {
	buffer []byte
	n      int
}

type MouseEvent struct {
	button int
	x      int
	y      int
	press  bool
}

type Event struct {
	kind       string
	keyEvent   KeyEvent
	mouseEvent MouseEvent
}

var terminalState TerminalState

func main() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}

	terminalState = initTerminalState()

	enterAltMode()
	enableMouseEvents()

	defer term.Restore(int(os.Stdin.Fd()), oldState)
	defer exitAltMode()
	defer disableMouseEvents()

	updateDimensions()
	draw(nil, 0)
	eventLoop()
}

func initTerminalState() TerminalState {
	return TerminalState{
		event: Event{},
		text:  make([]byte, 0, 128),
		dimensions: TerminalDimensions{
			width:  80,
			height: 25,
		},
	}
}

func enterAltMode() {
	fmt.Print("\x1b[?1049h")
	fmt.Print("\033[2J") // clear alt screen
	// fmt.Print("\033[H")  // move cursor to top-left
}

func enableMouseEvents() {
	fmt.Print("\033[?1000h") //enable mouse clicks
	fmt.Print("\033[?1003h") //enable motion tracking
	fmt.Print("\033[?1006h") //enable SGR mode
}

func draw(buf []byte, n int) {
	fmt.Print("\x1b[H")  // home
	fmt.Print("\x1b[2K") // clear the line
	fmt.Printf("Terminal dimensions: %d x %d\r\n", terminalState.dimensions.width, terminalState.dimensions.height)
	
	fmt.Print("\x1b[2K") // clear the line
	fmt.Printf("Read %d bytes Char: %q\r\n", n, string(buf))
}

func eventLoop() {
	resizeCh := make(chan os.Signal, 1)
	eventCh := make(chan Event, 1)
	signal.Notify(resizeCh, syscall.SIGWINCH)

	go func() {
		buf := make([]byte, 64)
		for {
			n, err := os.Stdin.Read(buf)

			if err != nil {
				fmt.Printf("Error: %v", err)
				close(eventCh)
				return
			}

			input := string(buf[:n])

			if event, ok := parseMouseEvent(input); ok {
				eventCh <- Event{
					kind: "mouse",
					mouseEvent: MouseEvent{
						button: event.button,
						x:      event.x,
						y:      event.y,
						press:  event.press,
					},
				}
				continue
			}
			eventCh <- Event{
				kind: "key",
				keyEvent: KeyEvent{
					buffer: buf[:n],
					n:      n,
				},
			}
		}
	}()

	for {
		select {
		case _, ok := <-resizeCh:
			if !ok {
				return
			}
			updateDimensions()
			draw(nil, 0)

		case event, ok := <-eventCh:
			if !ok {
				return
			}
			switch event.kind {
			case "key":
				if handleKeyEvent(event.keyEvent) {
					return
				}
			case "mouse":
				handleMouseEvent(event.mouseEvent)
			}
			draw(terminalState.text, len(terminalState.text))
		}
	}
}

func parseMouseEvent(input string) (MouseEvent, bool) {
	var event MouseEvent
	var keyPress rune

	n, err := fmt.Sscanf(input, "\x1b[<%d;%d;%d%c", &event.button, &event.x, &event.y, &keyPress)

	if err != nil || n != 4 || (keyPress != 'M' && keyPress != 'm') {
		return MouseEvent{}, false
	}

	if keyPress == 'M' {
		event.press = true
	} else {
		event.press = false
	}

	return event, true
}

func handleKeyEvent(event KeyEvent) bool {
	if len(event.buffer) > 0 && event.n > 0 {
		byt := event.buffer[0]
		if byt == 3 {
			return true
		}
		if byt == 127 && len(terminalState.text) > 0 {
			terminalState.text = terminalState.text[:len(terminalState.text)-1]
		} else if byt >= 32 && byt <= 126 {
			terminalState.text = append(terminalState.text, byt)
		}
	}
	return false
}

func handleMouseEvent(event MouseEvent) {
	fmt.Print("\x1b[3;1H") // row 3, column 1
	fmt.Print("\x1b[2K")   // clear current line
	fmt.Printf("Mouse data: key=%d x=%d y=%d pressed=%v", event.button, event.x, event.y, event.press)
}

func exitAltMode() {
	fmt.Print("\x1b[?1049l")
}

func disableMouseEvents() {
	fmt.Print("\033[?1000l")
	fmt.Print("\033[?1003l")
	fmt.Print("\033[?1006l")
}

func getTerminalDimensions() (int, int, error) {
	return term.GetSize(int(os.Stdout.Fd()))
}

func updateDimensions() {
	width, height, err := getTerminalDimensions()

	if err != nil {
		return
	}

	terminalState.dimensions.width = width
	terminalState.dimensions.height = height
}
