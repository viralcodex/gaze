package tui

func drawBox(el *Element, dc *DrawContext) error {
	rect := el.computedRect
	style := styleFor(el)

	cell := Cell{Fg: style.fgColor, Bg: style.bgColor}

	dc.fill(rect, ' ', cell)

	if el.Style.Border == Auto {
		dc.drawBorders(rect, style.BorderChars, cell)
	}

	return nil
}
