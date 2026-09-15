package tui

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

func enterAltMode() {
	fmt.Print(AltScreenEnter)
	clearAltScreen()
}

func enableMouseEvents() {
	fmt.Print(MouseClickEnable)  //enable mouse clicks
	fmt.Print(MouseMotionEnable) //enable motion tracking
	fmt.Print(MouseSGREnable)    //enable SGR mode
}

func exitAltMode() {
	fmt.Print(AltScreenExit)
}

func disableMouseEvents() {
	fmt.Print(MouseClickDisable)
	fmt.Print(MouseMotionDisable)
	fmt.Print(MouseSGRDisable)
}

func clearAltScreen() {
	fmt.Print(ClearScreen, CursorHome)
}

func hideCursorPointer() {
	fmt.Print(HideCursor)
}

func showCursorPointer() {
	fmt.Print(ShowCursor)
}

func setCursorShape(c Cursor) {
	fmt.Print(cursorEscape(c))
}

func beginSynchronizedUpdate() {
	fmt.Print(BeginSynchronizedUpdate)
}

func endSynchronizedUpdate() {
	fmt.Print(EndSynchronizedUpdate)
}

func getTerminalDimensions() (TerminalDimensions, error) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return TerminalDimensions{}, err
	}

	return TerminalDimensions{
		Width:  width,
		Height: height,
	}, nil
}
