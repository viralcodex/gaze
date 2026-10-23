package tui

func drawButton(button *Element, dc *DrawContext) error {
	rect := button.computedRect
	contentRect := button.contentRect
	labelText := button.Label
	leftPadding := (contentRect.W - len(labelText)) / 2

	style := styleFor(button)

	//remove overflow label text
	if len(labelText) > contentRect.W {
		labelText = labelText[:contentRect.W]
	}

	cell := Cell{Fg: style.fgColor, Bg: style.bgColor}

	dc.fill(rect, ' ', cell)

	for i, ch := range labelText {
		dc.set(contentRect.X+leftPadding+i, contentRect.Y, rune(ch), cell)
	}

	if button.Style.Border == Auto {
		dc.drawBorders(rect, style.BorderChars, cell)
	}

	return nil
}
