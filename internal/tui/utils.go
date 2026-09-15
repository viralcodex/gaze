package tui

func minContentSize(el *Element) Rect {
	switch el.kind {
	case Button, Text:
		return Rect{W: len(el.Label), H: 1}
	case Image:
		return Rect{W: 1, H: 1}
	case Input:
		return Rect{W: 1, H: 3}
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

// overengineered for now
func (d Direction) getDirectedPadding(el *Element) DirectionVals {
	if d == Row {
		return DirectionVals{
			main: []int{
				el.Style.Padding.Left + el.Style.Padding.Right,
			},
			cross: []int{
				el.Style.Padding.Top + el.Style.Padding.Bottom,
			},
		}
	}
	return DirectionVals{
		main: []int{
			el.Style.Padding.Top + el.Style.Padding.Bottom,
		},
		cross: []int{
			el.Style.Padding.Left + el.Style.Padding.Right,
		},
	}
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
