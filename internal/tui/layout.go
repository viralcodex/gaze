package tui

import (
	"fmt"
	"os"
)

var frame *Frame = &Frame{}
var hitMap *HitMap
var mouseState MouseState = MouseState{}

func initGrid(root *Element) {
	hitMap = &HitMap{
		Width:    root.Rect.W,
		Height:   root.Rect.H,
		Cells:    make([]uint32, root.Rect.W*root.Rect.H),
		Elements: []*Element{nil},
	}
}

func Render(root *Element) {
	initGrid(root)
	computeLayout(root, &Rect{X: 0, Y: 0})
	buildHitMap(root)
	drawLayout(root)

	err := frame.flush(os.Stdout)
	if err != nil {
		fmt.Printf("An error occured while rendering: %v", err)
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
		return Rect{W: 1, H: 1}
	}
}

func buildHitMap(el *Element) {
	if el == nil || !el.State.Visible {
		return
	}

	elementId := uint32(len(hitMap.Elements))
	hitMap.Elements = append(hitMap.Elements, el)

	computedRect := el.ComputedRect

	x0 := max(0, computedRect.X-1)
	y0 := max(0, computedRect.Y-1)
	x1 := min(hitMap.Width, x0+computedRect.W)
	y1 := min(hitMap.Height, y0+computedRect.H)

	for y := y0; y < y1; y++ {
		row := y * hitMap.Width
		for x := x0; x < x1; x++ {
			hitMap.Cells[row+x] = elementId
		}
	}

	for _, child := range el.Children {
		buildHitMap(child)
	}
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
