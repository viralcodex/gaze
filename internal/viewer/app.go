package viewer

import (
	"flag"
	"fmt"
	"os"
	"time"

	"gaze/internal/tui"
)

type TerminalState struct {
	Event      tui.Event
	Root       *tui.Element
	Dimensions tui.TerminalDimensions
	ImageData  ImageData
}

var terminalState TerminalState
var program *tui.Program

func Run() error {
	filePath, err := getFileArgs()
	if err != nil {
		return err
	}

	image, err := loadImage(filePath)
	if err != nil {
		return err
	}

	program = tui.NewProgram(
		func(size tui.TerminalDimensions) *tui.Element {
			terminalState.Dimensions = size
			return createLayout()
		},
		func(size tui.TerminalDimensions) {
			terminalState.Dimensions = size
		},
		func(d time.Duration) bool {
			return false
		},
	)

	//to be removed after need
	program.OnDebug = func(snapshot tui.DebugSnapshot) []*tui.Element {
		debugParsed.Label = "parsed: " + snapshot.Parsed
		debugBuffered.Label = "buffered: " + snapshot.Buffered
		debugMouse.Label = fmt.Sprintf("mouse: action=%d button=%d x=%d y=%d hit=%s hovered=%s pressed=%s clicked=%s focused=%s",
			snapshot.Mouse.Action, snapshot.Mouse.Button, snapshot.Mouse.X, snapshot.Mouse.Y,
			snapshot.Hit, snapshot.Hovered, snapshot.Pressed, snapshot.Clicked, snapshot.Focused)
		return []*tui.Element{debugParsed, debugBuffered, debugMouse}
	}

	terminalState.ImageData = image
	registerImage()

	return program.Run()
}

func getFileArgs() (string, error) {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s <image-path>\n", os.Args[0])
	}

	if flag.NArg() != 1 {
		flag.Usage()
		return "", fmt.Errorf("expected one image path only")
	}

	return flag.Arg(0), nil
}
