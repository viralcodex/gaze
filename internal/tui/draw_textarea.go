package tui

func drawTextArea(el *Element, dc *DrawContext) error {
	rect := el.computedRect
	contentRect := el.contentRect
	style := styleFor(el)

	cell := Cell{Fg: style.fgColor, Bg: style.bgColor}
	dc.fill(rect, ' ', cell)

	border := 0
	if el.Style.Border == Auto {
		border = 1
	}

	if border == 1 {
		dc.drawBorders(rect, style.BorderChars, cell)
	}

	setVisualLines(el)
	text := el.inputState.Text

	// Draw the visual lines of the text area within the content rectangle
	// row, col := el.inputState.getVisualLine(el.visualLines)

	// for i := 0; i < len(el.visualLines); i++ {
	if len(el.visualLines) == 0 {
		return nil
	}

	for i, vl := range el.visualLines {
		row := contentRect.Y + i - el.inputState.Yoffset
		if row >= contentRect.Y+contentRect.H {
			break
		}
		rowContent := text[vl.Line][vl.startCol:vl.endCol]
		for colOffset, ch := range rowContent {
			col := contentRect.X + colOffset
			if col >= contentRect.X+contentRect.W || row < contentRect.Y {
				break
			}
			dc.set(col, row, ch, cell)
		}
	}

	return nil
}
