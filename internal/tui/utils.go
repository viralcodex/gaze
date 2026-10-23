package tui

func minContentSize(el *Element) Rect {
	switch el.kind {
	case Button, Text:
		return Rect{W: len(el.Label), H: 1}
	case Image:
		return Rect{W: 1, H: 1}
	case Input:
		return Rect{W: 1, H: 3}
	case TextArea:
		return Rect{W: 1, H: 5}
	default:
		return Rect{W: 1, H: 1}
	}
}

func (d Direction) getMain(rect *Rect) int {
	if d == Row {
		return rect.W
	}
	return rect.H
}

func (d Direction) setMain(rect *Rect, val int) {
	if d == Row {
		rect.W = val
		return
	}
	rect.H = val
}

func (d Direction) getCross(rect *Rect) int {
	if d == Row {
		return rect.H
	}
	return rect.W
}

func (d Direction) setCross(rect *Rect, val int) {
	if d == Row {
		rect.H = val
		return
	}
	rect.W = val
}

func (d Direction) getMainPos(r *Rect) int {
	if d == Row {
		return r.X
	}
	return r.Y
}

func (d Direction) setMainPos(rect *Rect, val int) {
	if d == Row {
		rect.X = val
		return
	}
	rect.Y = val
}

func (d Direction) getCrossPos(r *Rect) int {
	if d == Row {
		return r.Y
	}
	return r.X
}

func (d Direction) setCrossPos(rect *Rect, val int) {
	if d == Row {
		rect.Y = val
		return
	}
	rect.X = val
}

// getDirectedPadding returns el's padding summed along the parent's main and
// cross axes.
func (d Direction) getDirectedPadding(el *Element) (main, cross int) {
	p := el.Style.Padding
	if d == Row {
		return p.Left + p.Right, p.Top + p.Bottom
	}
	return p.Top + p.Bottom, p.Left + p.Right
}

// margin is a child's outer spacing, so the parent reserves it while
// aggregating. getMainMargin/getCrossMargin return the leading+trailing margin
// of el along the parent's main/cross axis.
func (d Direction) getMainMargin(el *Element) int {
	if d == Row {
		return el.Style.Margin.Left + el.Style.Margin.Right
	}
	return el.Style.Margin.Top + el.Style.Margin.Bottom
}

func (d Direction) getCrossMargin(el *Element) int {
	if d == Row {
		return el.Style.Margin.Top + el.Style.Margin.Bottom
	}
	return el.Style.Margin.Left + el.Style.Margin.Right
}

// lead margin is the leading edge only (left/top), used while advancing the
// placement cursor; the trailing margin is the total minus the lead.
func (d Direction) getMainMarginLead(el *Element) int {
	if d == Row {
		return el.Style.Margin.Left
	}
	return el.Style.Margin.Top
}

func (d Direction) getCrossMarginLead(el *Element) int {
	if d == Row {
		return el.Style.Margin.Top
	}
	return el.Style.Margin.Left
}

func (d Direction) getMeasuredRect(main, cross int) Rect {
	var measuredRect Rect

	if d == Row {
		measuredRect.W = main
		measuredRect.H = cross
	} else {
		measuredRect.H = main
		measuredRect.W = cross
	}

	return measuredRect
}

func (m *HitMap) at(x, y int) *Element {
	if x < 0 || y < 0 || x >= m.Width || y >= m.Height {
		return nil
	}

	id := m.Cells[y*m.Width+x]

	if id == 0 {
		return nil
	}

	return m.Elements[id]
}

func (dc *DrawContext) drawBorders(rect Rect, borders BorderChars, cell Cell) error {
	x0, y0 := rect.X, rect.Y
	x1, y1 := rect.X+rect.W-1, rect.Y+rect.H-1

	dc.set(x0, y0, borders.TopLeft, cell)
	dc.set(x1, y0, borders.TopRight, cell)
	dc.set(x0, y1, borders.BottomLeft, cell)
	dc.set(x1, y1, borders.BottomRight, cell)

	for x := x0 + 1; x < x1; x++ {
		dc.set(x, y0, borders.Top, cell)
		dc.set(x, y1, borders.Bottom, cell)
	}

	for y := y0 + 1; y < y1; y++ {
		dc.set(x0, y, borders.Left, cell)
		dc.set(x1, y, borders.Right, cell)
	}

	return nil
}

func (s *InputState) mapCursorPos(el *Element) (row, col int) {
	contentRect := el.contentRect
	visualLines := el.visualLines

	//for input which has no visualLines and only one row, place the cursor in the middle of the content rect
	if el.kind == Input {
		return contentRect.Y + contentRect.H/2, contentRect.X + s.Col - s.Xoffset
	}

	row, col = s.getVisualLine(visualLines)

	return contentRect.Y + row - s.Yoffset, contentRect.X + col
}

func setVisualLines(el *Element) {
	text := el.inputState.Text
	width := el.contentRect.W

	if width <= 0 {
		el.visualLines = nil
		return
	}

	var visualLines []visualLine

	//map lines to virtual lines and populate virtualLines
	for index, line := range text {
		if len(line) <= 0 {
			visualLines = append(visualLines, visualLine{Line: index, startCol: 0, endCol: 0})
			continue
		}
		for start := 0; start < len(line); start += width {
			end := min(start+width, len(line))
			visualLines = append(visualLines, visualLine{Line: index, startCol: start, endCol: end})
		}
	}
	el.visualLines = visualLines
}

func (s *InputState) getVisualLine(visualLines []visualLine) (row, col int) {
	if len(visualLines) == 0 {
		return 0, 0
	}
	last := 0
	for i, vl := range visualLines {
		if vl.Line != s.Line {
			continue
		}
		last = i
		if s.Col < vl.endCol {
			return i, s.Col - vl.startCol
		}
	}
	// s.Col sits at the end of its logical line: return that line's last
	// segment, not the globally last row.
	return last, s.Col - visualLines[last].startCol
}
