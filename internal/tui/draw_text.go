package tui

func drawText(el *Element, dc *DrawContext) error {
	style := styleFor(el)
	cell := Cell{Fg: style.fgColor, Bg: style.bgColor}
	dc.fill(el.computedRect, ' ', cell)

	x := el.contentRect.X
	for _, r := range el.Label {
		if x >= el.contentRect.X+el.contentRect.W {
			break
		}
		dc.set(x, el.contentRect.Y, r, cell)
		x++
	}
	return nil
}
