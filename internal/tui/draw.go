package tui

import (
	"fmt"
)

func drawButton(button *Element, c *Canvas) error {
	rect := button.computedRect
	contentRect := button.contentRect
	labelText := button.Label
	leftPadding := (contentRect.W - len(labelText)) / 2

	style := resolveStyle(button)

	//remove overflow label text
	if len(labelText) > contentRect.W {
		labelText = labelText[:contentRect.W]
	}

	cell := Cell{Fg: style.fgColor, Bg: style.bgColor}

	c.buff.fill(rect, ' ', cell)

	for i, ch := range labelText {
		c.buff.set(contentRect.X+leftPadding+i, contentRect.Y, rune(ch), cell)
	}

	if button.Style.Border == Auto {
		c.drawBorders(rect, style.BorderChars, cell)
	}

	return nil
}

func drawImage(el *Element, c *Canvas) error {
	outer := el.computedRect
	contentRect := el.contentRect
	style := el.Style

	cols, rows := c.fitToRect(el)

	x := contentRect.X + alignOffset(el.ImageRef.Position.X, contentRect.W-cols)
	y := contentRect.Y + alignOffset(el.ImageRef.Position.Y, contentRect.H-rows)

	currentImage, ok := c.imageData[el.ImageRef.ID]

	if !ok {
		return fmt.Errorf("%s", "No image found")
	}

	if currentImage.NeedsUpload {
		if err := c.uploadImage(currentImage); err != nil {
			return err
		}
	}

	c.buff.fill(outer, ' ', Cell{Bg: style.bgColor})

	if err := c.placeImage(currentImage, x, y, cols, rows); err != nil {
		return err
	}
	return nil
}

// alignOffset places a fitted span of length `free` (box - image) along one
// axis. Shared with the flex leading-offset logic in justifyOffsets.
func alignOffset(a Align, free int) int {
	if free <= 0 {
		return 0
	}
	switch a {
	case AlignCenter:
		return free / 2
	case AlignEnd:
		return free
	}
	return 0 // AlignStart
}

func drawBox(el *Element, c *Canvas) error {
	rect := el.computedRect
	style := resolveStyle(el)

	cell := Cell{Fg: style.fgColor, Bg: style.bgColor}

	c.buff.fill(rect, ' ', cell)

	if el.Style.Border == Auto {
		c.drawBorders(rect, style.BorderChars, cell)
	}

	return nil
}

func drawText(el *Element, c *Canvas) error {
	style := resolveStyle(el)
	cell := Cell{Fg: style.fgColor, Bg: style.bgColor}
	c.buff.fill(el.computedRect, ' ', cell)

	x := el.contentRect.X
	for _, r := range el.Label {
		if x >= el.contentRect.X+el.contentRect.W {
			break
		}
		c.buff.set(x, el.contentRect.Y, r, cell)
		x++
	}
	return nil
}

func drawInput(el *Element, c *Canvas) error {
	rect := el.computedRect
	style := resolveStyle(el)

	cell := Cell{Fg: style.fgColor, Bg: style.bgColor}
	c.buff.fill(rect, ' ', cell)

	border := 0
	if el.Style.Border == Auto {
		border = 1
	}

	if border == 1 {
		c.drawBorders(rect, style.BorderChars, cell)
	}

	text := el.inputState.Text[el.inputState.ScrollOffset:]
	limit := el.contentRect.X + el.contentRect.W

	if len(text) == 0 && !el.state.Focused {
		text = []rune(el.Placeholder)
	}

	x := el.contentRect.X
	for _, r := range text {
		if x >= limit {
			break
		}
		c.buff.set(x, el.contentRect.Y+(el.contentRect.H/2), r, cell)
		x++
	}

	return nil
}

func (c *Canvas) drawBorders(rect Rect, borders BorderChars, cell Cell) error {
	x0, y0 := rect.X, rect.Y
	x1, y1 := rect.X+rect.W-1, rect.Y+rect.H-1

	c.buff.set(x0, y0, borders.TopLeft, cell)
	c.buff.set(x1, y0, borders.TopRight, cell)
	c.buff.set(x0, y1, borders.BottomLeft, cell)
	c.buff.set(x1, y1, borders.BottomRight, cell)

	for x := x0 + 1; x < x1; x++ {
		c.buff.set(x, y0, borders.Top, cell)
		c.buff.set(x, y1, borders.Bottom, cell)
	}

	for y := y0 + 1; y < y1; y++ {
		c.buff.set(x0, y, borders.Left, cell)
		c.buff.set(x1, y, borders.Right, cell)
	}

	return nil
}
