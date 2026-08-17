package tui

import (
	"fmt"
	"strings"
)

func drawButton(button *Element) {
	outer := button.ComputedRect
	content := button.ContentRect
	labelText := button.Label

	//remove overflow label text
	if len(labelText) > content.W {
		labelText = labelText[:content.W]
	}

	leftPadding := (content.W - len(labelText)) / 2
	rightPadding := content.W - len(labelText) - leftPadding

	label := strings.Repeat(" ", leftPadding) + labelText + strings.Repeat(" ", rightPadding)

	style := resolveStyle(button)

	if button.Style.Border == None {
		frame.writeOut(
			cursorPosition(content.Y, content.X),
			foregroundColor(style.fgColor),
			backgroundColor(style.bgColor),
			label,
			ResetStyle,
		)
		return
	}

	borderWidth := max(0, outer.W-2)
	topBorder := strings.Repeat(string(style.BorderChars.Top), borderWidth)
	bottomBorder := strings.Repeat(string(style.BorderChars.Bottom), borderWidth)

	labelRow := content.Y

	frame.writeOut(
		cursorPosition(outer.Y, outer.X),
		foregroundColor(style.fgColor),
		backgroundColor(style.bgColor),
		string(style.BorderChars.TopLeft),
		topBorder,
		string(style.BorderChars.TopRight),
		ResetStyle,
	)

	for r := outer.Y + 1; r < outer.Y+outer.H-1; r++ {
		line := strings.Repeat(" ", outer.W-2)

		if r == labelRow {
			labelOffset := content.X - (outer.X + 1)
			line = line[:labelOffset] + label + line[labelOffset+len(label):]
		}

		frame.writeOut(
			cursorPosition(r, outer.X),
			foregroundColor(style.fgColor),
			backgroundColor(style.bgColor),
			string(style.BorderChars.Left),
			line,
			string(style.BorderChars.Right),
			ResetStyle,
		)
	}

	frame.writeOut(
		cursorPosition(outer.Y+outer.H-1, outer.X),
		foregroundColor(style.fgColor),
		backgroundColor(style.bgColor),
		string(style.BorderChars.BottomLeft),
		bottomBorder,
		string(style.BorderChars.BottomRight),
		ResetStyle,
	)
}

func drawImage(el *Element) {
	rect := el.ContentRect
	cols, rows := FitToRect(rect)

	if currentImage.NeedsUpload {
		if err := UploadImageData(cols, rows); err != nil {
			return
		}
	}

	PlaceImage(rect.X, rect.Y, cols, rows)
}

func drawBox(el *Element) {
}

func drawInput(el *Element) {
}

func cursorPosition(row, col int) string {
	return fmt.Sprintf(CursorPosition, row, col)
}
