package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

type TerminalDimensions struct {
	width  int
	height int
}

var dimensions = TerminalDimensions{
	width:  80,
	height: 25,
}

func main() {
	enterAltMode()
	defer exitAltMode()

	updateDimensions()
	draw()
	eventLoop()
}

func enterAltMode() {
	fmt.Print("\x1b[?1049h")
	fmt.Print("\033[2J") // clear alt screen
	fmt.Print("\033[H")  // move cursor to top-left
}

func draw() {
	fmt.Print("\x1b7")       // save the cursor position
	defer fmt.Print("\x1b8") // restore the cursor position

	fmt.Print("\x1b[H")
	fmt.Print("\x1b[2K")

	fmt.Printf("Terminal dimensions: %d x %d", dimensions.width, dimensions.height)
}

func eventLoop() {
	onResize()
	for {
	}
}

func exitAltMode() {
	fmt.Print("\x1b[?1049l")
}

func getTerminalDimensions() (int, int, error) {
	return term.GetSize(int(os.Stdout.Fd()))
}

func onResize() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGWINCH)

	go func() {
		for {
			if _, ok := <-sig; !ok {
				return
			}

			updateDimensions()
			draw()
		}
	}()
}

func updateDimensions() {
	width, height, err := getTerminalDimensions()

	if err != nil {
		return
	}

	dimensions.width = width
	dimensions.height = height
}
