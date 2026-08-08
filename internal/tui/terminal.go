package tui

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

func EnterAltMode() {
	fmt.Print(AltScreenEnter)
	ClearAltScreen()
}

func EnableMouseEvents() {
	fmt.Print(MouseClickEnable)  //enable mouse clicks
	fmt.Print(MouseMotionEnable) //enable motion tracking
	// fmt.Print(MouseSGREnable)    //enable SGR mode
}

func ExitAltMode() {
	fmt.Print(AltScreenExit)
}

func DisableMouseEvents() {
	fmt.Print(MouseClickDisable)
	fmt.Print(MouseMotionDisable)
	fmt.Print(MouseSGRDisable)
}

func ClearAltScreen() {
	fmt.Print("\033[4;1H")
	fmt.Print("\033[0J")
	fmt.Print(ClearScreen)
}

func GetTerminalDimensions() (TerminalDimensions, error) {
    width, height, err := term.GetSize(int(os.Stdout.Fd()))
    if err != nil {
        return TerminalDimensions{}, err
    }

    return TerminalDimensions{
        Width:  width,
        Height: height,
    }, nil
}
