package tui

func (p *Program) render(el *Element) error {
	p.measureLayout(el)
	p.arrangeLayout(el)
	p.rebuildHitMap(el)

	p.Canvas.buff.fill(Rect{W: p.Dimensions.Width, H: p.Dimensions.Height}, ' ', Cell{})

	return p.drawFrame(el)
}

func (p *Program) drawFrame(el *Element) error {
	if err := p.drawElements(el); err != nil {
		return err
	}
	return p.flushFrame()
}

func (p *Program) paintDirtyElements(elements []*Element) error {
	for _, el := range elements {
		if err := p.drawElements(el); err != nil {
			return err
		}
	}

	return p.flushFrame()
}

// bottom up element size measurement (explicit Rect takes top predence and overrides measured values)
func (p *Program) measureLayout(el *Element) {
	for _, child := range el.children {
		p.measureLayout(child)
	}

	dir := el.Layout.Direction

	var mainMax, crossMax, n, border int

	for _, child := range el.children {
		if child.Style.Position == Absolute { //absolute elements don't participate in parent layouting
			continue
		}
		mainMax += dir.getMain(&child.measuredRect)
		crossMax = max(crossMax, dir.getCross(&child.measuredRect))
		n++
	}

	//if no children (leaf nodes), assign minimum defaults
	if n == 0 {
		contentSize := minContentSize(el)
		mainMax = dir.getMain(&contentSize)
		crossMax = dir.getCross(&contentSize)
	}

	//add gaps between children
	mainMax += el.Layout.Gap * (max(1, n) - 1)

	dirPadding := dir.getDirectedPadding(el)

	if el.Style.Border == Auto {
		border = 1
	}

	finalMain := mainMax + dirPadding.main[0] + 2*border
	finalCross := crossMax + dirPadding.cross[0] + 2*border

	el.measuredRect = dir.getMeasuredRect(finalMain, finalCross)

	//override with Rect values (if provided)
	if el.Rect.W != 0 {
		el.measuredRect.W = el.Rect.W
	}
	if el.Rect.H != 0 {
		el.measuredRect.H = el.Rect.H
	}
}

// computing element placements
func (p *Program) arrangeLayout(el *Element) {
	if el == nil {
		return
	}
	el.contentRect = getContentRect(el)
	p.arrangeChildren(el)
	for _, child := range el.children {
		p.arrangeLayout(child)
	}
}

func (p *Program) drawElements(el *Element) error {
	if el == nil || !el.state.Visible {
		return nil
	}

	if el.Draw != nil {
		if err := el.Draw(el, &p.Canvas); err != nil {
			return err
		}
	}

	for _, child := range el.children {
		if err := p.drawElements(child); err != nil {
			return err
		}
	}
	return nil
}

func (p *Program) rebuildHitMap(el *Element) {
	clear(p.hitMap.Cells)
	p.hitMap.Elements = p.hitMap.Elements[:1]
	p.buildhitMap(el)
}

func (p *Program) buildhitMap(el *Element) {
	if el == nil || !el.state.Visible {
		return
	}

	elementId := uint32(len(p.hitMap.Elements))
	p.hitMap.Elements = append(p.hitMap.Elements, el)

	computedRect := el.computedRect

	x0 := max(0, computedRect.X)
	y0 := max(0, computedRect.Y)
	x1 := min(p.hitMap.Width, x0+computedRect.W)
	y1 := min(p.hitMap.Height, y0+computedRect.H)

	for y := y0; y < y1; y++ {
		row := y * p.hitMap.Width
		for x := x0; x < x1; x++ {
			p.hitMap.Cells[row+x] = elementId
		}
	}

	for _, child := range el.children {
		p.buildhitMap(child)
	}
}

// this places el.children inside el.contentRect (usable space)
func (p *Program) arrangeChildren(el *Element) {
	dir := el.Layout.Direction

	var totalMain, usedMain, freeMain, totalGrow, n, start, lead, gap int

	for _, child := range el.children {
		if child.Style.Position == Absolute {
			continue // out of flow, don't count it (same as measure)
		}
		usedMain += dir.getMain(&child.measuredRect)
		totalGrow += child.Layout.Grow
		n++
	}

	totalMain = dir.getMain(&el.contentRect)
	usedMain += el.Layout.Gap * (max(1, n) - 1)
	freeMain = max(0, totalMain-usedMain)

	// either distribute free space to every child having grow > 0 in order of weights
	// or use it to justify based on user param (default is start)
	if totalGrow == 0 {
		lead, gap = justifyOffsets(el.Layout.Justify, freeMain, n)
	}

	start = dir.getMainPos(&el.contentRect)
	mainPos := start + lead

	for _, child := range el.children {
		if child.Style.Position == Absolute {
			continue
		}
		childMain := dir.getMain(&child.measuredRect)
		if totalGrow > 0 {
			childGrow := child.Layout.Grow
			share := (freeMain * childGrow) / totalGrow
			childMain += share
		}
		dir.setMain(&child.computedRect, childMain)
		dir.setMainPos(&child.computedRect, mainPos)

		placeCross(el, child)

		mainPos += childMain + gap + el.Layout.Gap //move position ahead
	}

	// absolute children sit out of flow: positioned against the parent's
	// content origin using their own Rect offset, sized from measuredRect
	for _, child := range el.children {
		if child.Style.Position != Absolute {
			continue
		}
		child.computedRect = Rect{
			X: el.contentRect.X + child.Rect.X,
			Y: el.contentRect.Y + child.Rect.Y,
			W: child.measuredRect.W,
			H: child.measuredRect.H,
		}
	}
}

func justifyOffsets(justify Justify, free int, n int) (int, int) {
	switch justify {
	case Center:
		return free / 2, 0
	case End:
		return free, 0
	case SpaceEvenly:
		return free / (n + 1), free / (n + 1)
	case SpaceAround:
		return free / (2 * n), free / n
	case SpaceBetween:
		if n == 1 {
			return 0, 0
		}
		return 0, free / (n - 1)
	}
	return 0, 0
}

func placeCross(el *Element, child *Element) {
	dir := el.Layout.Direction
	contentCross := dir.getCross(&el.contentRect)   // gives width/height
	crossPos := dir.getCrossPos(&el.contentRect)    // gives starting X or Y
	childCross := dir.getCross(&child.measuredRect) // gives width/height (of child)

	switch el.Layout.Align {
	case AlignCenter:
		crossPos += max(0, (contentCross-childCross)/2)
	case AlignEnd:
		crossPos += max(0, contentCross-childCross)
	case AlignStretch:
		childCross = contentCross
	}
	dir.setCrossPos(&child.computedRect, crossPos)
	dir.setCross(&child.computedRect, childCross)
}

func getContentRect(el *Element) Rect {
	border := 0
	if el.Style.Border == Auto {
		border = 1
	}

	padding := el.Style.Padding

	return Rect{
		X: el.computedRect.X + border + padding.Left,
		Y: el.computedRect.Y + border + padding.Top,
		W: max(0, el.computedRect.W-2*border-padding.Left-padding.Right),
		H: max(0, el.computedRect.H-2*border-padding.Top-padding.Bottom),
	}
}
