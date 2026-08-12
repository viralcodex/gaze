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
	outer := button.ComputedRect
	content := button.ContentRect

	// innerWidth := max(0, outer.W-2)
	labelText := button.Label

	//remove excess label text
	if len(labelText) > content.W {
		labelText = labelText[:content.W]
	}

	leftPadding := (content.W - len(labelText)) / 2
	rightPadding := content.W - len(labelText) - leftPadding

	label := strings.Repeat(" ", leftPadding) + labelText + strings.Repeat(" ", rightPadding)

	if button.Style.Border == None {
		fmt.Printf("%s%s%s%s%s",
			cursorPosition(content.Y, content.X),
			foregroundColor(button.Style.fgColor),
			backgroundColor(button.Style.bgColor),
			label,
			ResetStyle,
		)
		return
	}

	borderWidth := max(0, outer.W-2)
	topBorder := strings.Repeat(string(button.Style.BorderChars.Top), borderWidth)
	bottomBorder := strings.Repeat(string(button.Style.BorderChars.Bottom), borderWidth)

	labelRow := content.Y

	fmt.Printf("%s%s%s%s%s%s%s",
		cursorPosition(outer.Y, outer.X),
		foregroundColor(button.Style.fgColor),
		backgroundColor(button.Style.bgColor),
		string(button.Style.BorderChars.TopLeft),
		topBorder,
		string(button.Style.BorderChars.TopRight),
		ResetStyle,
	)

	for r := outer.Y + 1; r < outer.Y+outer.H-1; r++ {
		line := strings.Repeat(" ", outer.W-2)

		if r == labelRow {
			labelOffset := content.X - (outer.X + 1)
			line = line[:labelOffset] + label + line[labelOffset+len(label):]
		}

		fmt.Printf(
			"%s%s%s%s%s%s%s",
			cursorPosition(r, outer.X),
			foregroundColor(button.Style.fgColor),
			backgroundColor(button.Style.bgColor),
			string(button.Style.BorderChars.Left),
			line,
			string(button.Style.BorderChars.Right),
			ResetStyle,
		)
	}

	fmt.Printf("%s%s%s%s%s%s%s",
		cursorPosition(outer.Y+outer.H-1, outer.X),
		foregroundColor(button.Style.fgColor),
		backgroundColor(button.Style.bgColor),
		string(button.Style.BorderChars.BottomLeft),
		bottomBorder,
		string(button.Style.BorderChars.BottomRight),
		ResetStyle,
	)
}

func drawImage(el *Element) {
	rect := el.ContentRect
	cols, rows := FitToRect(rect)

	if currentImage.NeedsUpload {
		fmt.Print(cursorPosition(rect.Y, rect.X))
		UploadImageData(cols, rows)
		return
	}

	PlaceImage(rect.X, rect.Y, cols, rows)
}

func drawBox(el *Element) {
}

func drawInput(el *Element) {
}

func cursorPosition(row, col int) string {
	return fmt.Sprintf("%s%d;%dH", CSI, row, col)
}
