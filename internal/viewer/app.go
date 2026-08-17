package viewer

import (
	"flag"
	"fmt"
	"os"
	"os/signal"

	// "strconv"
	// "strings"

	"syscall"

	"gaze/internal/tui"

	"golang.org/x/term"
)

type TerminalState struct {
	Event      tui.Event
	Dimensions tui.TerminalDimensions
	Root       *tui.Element
	Focused    *tui.Element
}

var terminalState TerminalState
var img Image

func Run() error {
	filePath, err := getFileArgs()
	if err != nil {
		return err
	}

	image, err := LoadImage(filePath)
	if err != nil {
		return err
	}

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	if err != nil {
		return err
	}

	tui.EnterAltMode()
	defer tui.ExitAltMode()

	tui.HideCursorPointer()
	defer tui.ShowCursorPointer()

	tui.EnableMouseEvents()
	defer tui.DisableMouseEvents()

	initViewerState(image)
	// updateTerminalDimensions()

	sendImageData()
	getImageRect()
	createLayout()

	requestRender()
	eventLoop()

	return nil
}

func initViewerState(image Image) {
	terminalState = TerminalState{
		Event:      tui.Event{},
	}

	updateTerminalDimensions()

	img = image

	terminalState.Root = tui.NewBox("root", tui.Rect{
			X: 1,
			Y: 1,
			W: terminalState.Dimensions.Width,
			H: terminalState.Dimensions.Height,
		}, tui.Style{},
	)
}

func eventLoop() {
	resizeCh := make(chan os.Signal, 1)
	eventCh := make(chan tui.Event, 1)
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

			if event, ok := tui.ParseMouseEvent(input); ok {
				eventCh <- tui.Event{
					Kind: "mouse",
					MouseEvent: tui.MouseEvent{
						Button: event.Button,
						X:      event.X,
						Y:      event.Y,
						Action: event.Action,
					},
				}
				continue
			}
			eventCh <- tui.Event{
				Kind: "key",
				KeyEvent: tui.KeyEvent{
					Buffer: buf[:n],
					N:      n,
				},
			}
		}
	}()

	for {
		isDirty := false
		select {
		case _, ok := <-resizeCh:
			if !ok {
				return
			}
			updateLayout()
			tui.ClearAltScreen()
			tui.MarkImageReupload()
			isDirty = true

		case event, ok := <-eventCh:
			if !ok {
				return
			}
			switch event.Kind {
			case "key":
				if tui.HandleKeyEvent(event.KeyEvent) {
					return
				}
			case "mouse":
				isDirty = tui.HandleMouseEvent(event.MouseEvent)
			}
		}
		if isDirty {
			requestRender()
		}
	}
}

func getFileArgs() (string, error) {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s <image-path>\n", os.Args[0])
	}

	// flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		return "", fmt.Errorf("expected one image path only")
	}

	return flag.Arg(0), nil
}

func requestRender() {
	tui.Render(terminalState.Root)
}
