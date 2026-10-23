package tui

func (p *Program) render(el *Element) error {
	p.measureLayout(el)
	p.arrangeLayout(el)
	p.resolveStyles(el, nil)
	p.rebuildHitMap(el)

	p.Canvas.buff.reset()

	return p.drawFrame(el)
}

// resolveStyles caches each element's base style with parent inheritance folded
// in, top-down so a parent is resolved before its children read from it. Only
// Fg/Bg cascade; state overrides stay local and are applied by styleFor at
// draw time. Runs on full renders only, so partial repaints reuse the cache.
func (p *Program) resolveStyles(el *Element, parent *Style) {
	if el == nil {
		return
	}

	style := el.Style
	if parent != nil {
		if style.Fg == "" {
			style.Fg = parent.Fg
			style.fgColor = parent.fgColor
		}
		if style.Bg == "" {
			style.Bg = parent.Bg
			style.bgColor = parent.bgColor
		}
	}
	el.resolvedStyle = style

	for _, child := range el.children {
		p.resolveStyles(child, &el.resolvedStyle)
	}
}

func (p *Program) drawFrame(el *Element) error {
	if err := p.drawElements(el, p.Root.computedRect); err != nil {
		return err
	}
	return p.flushFrame()
}

func (p *Program) paintDirtyElements() error {
	for _, el := range p.dirtyElements {
		if err := p.drawElements(el, p.clipForElement(el)); err != nil {
			return err
		}
	}
	p.dirtyElements = p.dirtyElements[:0]
	return p.flushFrame()
}

// clipForElement reconstructs the clip an element would receive during a full
// draw, by walking up and intersecting the content rect of every container above it.
func (p *Program) clipForElement(el *Element) Rect {
	clip := Rect{W: p.Dimensions.Width, H: p.Dimensions.Height}
	for a := el.Parent; a != nil; a = a.Parent {
		if a.Layout.Overflow == OverflowHidden {
			clip = clipRect(clip, a.contentRect)
		}
	}
	return clip
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
		mainMax += dir.getMain(&child.measuredRect) + dir.getMainMargin(child)
		crossMax = max(crossMax, dir.getCross(&child.measuredRect)+dir.getCrossMargin(child))
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

	dirPaddingMain, dirPaddingCross := dir.getDirectedPadding(el)

	if el.Style.Border == Auto {
		border = 1
	}

	finalMain := mainMax + dirPaddingMain + 2*border
	finalCross := crossMax + dirPaddingCross + 2*border

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

func (p *Program) drawElements(el *Element, clip Rect) error {
	if el == nil || !el.state.Visible {
		return nil
	}

	if el.Draw != nil {
		dc := DrawContext{clip: clip, Canvas: &p.Canvas}
		if err := el.Draw(el, &dc); err != nil {
			return err
		}
	}

	childClip := clip
	if el.Layout.Overflow == OverflowHidden {
		childClip = clipRect(clip, el.contentRect)
	}

	for _, child := range el.children {
		if err := p.drawElements(child, childClip); err != nil {
			return err
		}
	}
	return nil
}

func clipRect(clip Rect, parentRect Rect) Rect {
	x0 := max(clip.X, parentRect.X)
	y0 := max(clip.Y, parentRect.Y)
	x1 := min(clip.X+clip.W, parentRect.X+parentRect.W)
	y1 := min(clip.Y+clip.H, parentRect.Y+parentRect.H)

	if x0 >= x1 || y0 >= y1 {
		return Rect{}
	}

	return Rect{
		X: x0,
		Y: y0,
		W: x1 - x0,
		H: y1 - y0,
	}
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
		usedMain += dir.getMain(&child.measuredRect) + dir.getMainMargin(child)
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
		mainMargin := dir.getMainMargin(child)
		leadMargin := dir.getMainMarginLead(child)

		mainPos += leadMargin
		dir.setMain(&child.computedRect, childMain)
		dir.setMainPos(&child.computedRect, mainPos)

		placeCross(el, child)

		// advance past the child, its trailing margin, then the gaps
		mainPos += childMain + (mainMargin - leadMargin) + gap + el.Layout.Gap
	}

	// absolute children sit out of flow: positioned against the parent's
	// content origin using their own Rect offset, sized from measuredRect
	for _, child := range el.children {
		if child.Style.Position != Absolute {
			continue
		}
		child.computedRect = Rect{
			X: el.contentRect.X + child.Rect.X + child.Style.Margin.Left,
			Y: el.contentRect.Y + child.Rect.Y + child.Style.Margin.Top,
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
	crossStart := dir.getCrossPos(&el.contentRect)  // gives starting X or Y
	childCross := dir.getCross(&child.measuredRect) // gives width/height (of child)

	crossMargin := dir.getCrossMargin(child)
	leadMargin := dir.getCrossMarginLead(child)
	trailMargin := crossMargin - leadMargin

	crossPos := crossStart + leadMargin // AlignStart default

	switch el.Layout.Align {
	case AlignCenter:
		crossPos = crossStart + leadMargin + max(0, (contentCross-childCross-crossMargin)/2)
	case AlignEnd:
		crossPos = crossStart + max(0, contentCross-childCross-trailMargin)
	case AlignStretch:
		childCross = max(0, contentCross-crossMargin)
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

func (p *Program) MarkDirty(el ...*Element) {
	p.dirtyElements = append(p.dirtyElements, el...)
}

func (p *Program) RequestRender() {
	p.isLayoutDirty = true
}

func (p *Program) SetSelected(el *Element, selected bool) {
	if el == nil || el.state.Selected == selected {
		return
	}
	el.state.Selected = selected
	p.MarkDirty(el)
}

func (p *Program) Post(fn func()) {
	p.postCh <- fn
}
