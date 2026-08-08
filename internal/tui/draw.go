package tui

import (
	"fmt"
	"strings"
)

// func draw(buf []byte, n int) {
// 	fmt.Print(CursorHome) // home
// 	fmt.Print(ClearLine)  // clear the line
// 	// fmt.Printf("Terminal dimensions: %d x %d\r\n", viewer.TerminalState.Dimensions.Width, terminalState.Dimensions.Height)
// 	fmt.Print(ClearLine) // clear the line
// 	fmt.Printf("Read %d bytes Char: %q\r\n", n, string(buf))
// }

func drawButton(button *Element) {
	topBorder := strings.Repeat("─", button.Rect.W-2)
	bottomBorder := strings.Repeat("─", button.Rect.W-2)

	innerWidth := button.Rect.W - 2
	labelText := button.Label

	if len(labelText) > innerWidth {
		labelText = labelText[:innerWidth]
	}

	leftPadding := (innerWidth - len(labelText)) / 2
	rightPadding := innerWidth - len(labelText) - leftPadding
	label := strings.Repeat(" ", leftPadding) + labelText + strings.Repeat(" ", rightPadding)

	fmt.Printf("%s┌%s┐", cursorPosition(button.Rect.Y, button.Rect.X), topBorder)
	fmt.Printf("%s│%s│", cursorPosition(button.Rect.Y+1, button.Rect.X), label)
	fmt.Printf("%s└%s┘", cursorPosition(button.Rect.Y+2, button.Rect.X), bottomBorder)
}

func cursorPosition(row, col int) string {
	return fmt.Sprintf("%s%d;%dH", CSI, row, col)
}

func drawImage(el *Element) {
	cols, rows := FitToRect(el.Rect)

	if currentImage.NeedsUpload {
		fmt.Print(cursorPosition(el.Rect.Y, el.Rect.X))
		UploadImageData(cols, rows)
		return
	}

	PlaceImage(el.Rect.X, el.Rect.Y, cols, rows)
}
