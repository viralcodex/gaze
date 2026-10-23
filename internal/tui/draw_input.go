package tui

func drawInput(el *Element, dc *DrawContext) error {
	rect := el.computedRect
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

	text := el.inputState.Text[0]

	if len(text) == 0 {
		text = []rune(el.Placeholder)
	} else {
		//we need to update the offset on resize
		el.inputState.Xoffset = max(0, len(text)-el.contentRect.W)
		text = text[el.inputState.Xoffset:]
	}

	x := el.contentRect.X
	limit := x + el.contentRect.W

	for _, r := range text {
		if x >= limit {
			break
		}
		dc.set(x, el.contentRect.Y+(el.contentRect.H/2), r, cell)
		x++
	}

	return nil
}
