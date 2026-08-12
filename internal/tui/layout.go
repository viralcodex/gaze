package tui

func Render(root *Element) {
	computeLayout(root, &Rect{X: 0, Y: 0})
	drawLayout(root)
}

func drawLayout(el *Element) {
	if el == nil || !el.State.Visible {
		return
	}

	switch el.Kind {
	case ElementBox:
		drawBox(el)
	case ElementButton:
		drawButton(el)
	case ElementInput:
		drawInput(el)
	case ElementImage:
		drawImage(el)
	}
	for _, child := range el.Children {
		drawLayout(child)
	}
}

func computeLayout(el *Element, parentRect *Rect) {
	if el == nil {
		return
	}

	getComputedRect(el, parentRect)

	for _, child := range el.Children {
		computeLayout(child, &el.ContentRect)
	}
}

func getComputedRect(el *Element, parentRect *Rect) {
	margin := el.Style.Margin
	padding := el.Style.Padding
	border := 0

	if el.Style.Border == Auto {
		border = 1
	}

	minContentRect := minContentSize(el)

	w := el.Rect.W
	if w == 0 {
		w = minContentRect.W + 2*border + padding.Left + padding.Right
	}

	h := el.Rect.H
	if h == 0 {
		h = minContentRect.H + 2*border + padding.Top + padding.Bottom
	}

	switch el.Style.Position {

	case PositionAbsolute:
		el.ComputedRect = Rect{
			X: el.Rect.X + margin.Left,
			Y: el.Rect.Y + margin.Top,
			W: w,
			H: h,
		}

	case PositionRelative:
		el.ComputedRect = Rect{
			X: parentRect.X + el.Rect.X + margin.Left,
			Y: parentRect.Y + el.Rect.Y + margin.Top,
			W: w,
			H: h,
		}

	default:
		el.ComputedRect = Rect{
			X: parentRect.X + el.Rect.X,
			Y: parentRect.Y + el.Rect.Y,
			W: w,
			H: h,
		}
	}
	el.ContentRect = Rect{
		X: el.ComputedRect.X + border + padding.Left,
		Y: el.ComputedRect.Y + border + padding.Top,
		W: max(0, el.ComputedRect.W-2*border-padding.Left-padding.Right),
		H: max(0, el.ComputedRect.H-2*border-padding.Top-padding.Bottom),
	}
}

func minContentSize(el *Element) Rect {
	switch el.Kind {
	case ElementButton:
		return Rect{W: len(el.Label), H: 1}
	case ElementText:
		return Rect{W: len(el.Label), H: 1}
	case ElementImage:
		return Rect{W: 1, H: 1}
	default:
		return Rect{W: 0, H: 0}
	}
}
