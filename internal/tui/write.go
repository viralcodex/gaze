package tui

import (
	"io"
	"os"
	"strconv"
)

func (dc *DrawContext) fill(rect Rect, r rune, cell Cell) {
	for y := rect.Y; y < rect.Y+rect.H; y++ {
		for x := rect.X; x < rect.X+rect.W; x++ {
			dc.set(x, y, r, cell)
		}
	}
}

func (dc *DrawContext) set(x, y int, r rune, cell Cell) {
	dc.Canvas.buff.set(x, y, r, cell, dc.clip)
}

func (b *Buffer) set(x, y int, r rune, cell Cell, clipRect Rect) {
	if x < 0 || y < 0 || x >= b.Width || y >= b.Height {
		return
	}

	if x < clipRect.X || y < clipRect.Y || x >= clipRect.X+clipRect.W || y >= clipRect.Y+clipRect.H {
		return
	}

	cell.Rune = r
	b.Cells[y*b.Width+x] = cell
}

// resets the whole buffer null cells
func (b *Buffer) reset() {
	for i := range b.Cells {
		b.Cells[i] = Cell{}
	}
}

func (b *Buffer) at(x, y int) *Cell {
	return &b.Cells[y*b.Width+x]
}

func (p *Program) flushFrame() error {
	var lastFg, lastBg Color
	lastX, lastY := -2, -2
	canvas := p.Canvas

	for y := 0; y < canvas.buff.Height; y++ {
		for x := 0; x < canvas.buff.Width; x++ {
			b := canvas.buff.at(x, y)
			if *b == *p.frontbuff.at(x, y) {
				continue
			}
			if x != lastX+1 || y != lastY {
				canvas.frame.setCursorPosition(y, x)
			}
			if b.Fg != lastFg || b.Bg != lastBg {
				canvas.frame.writeTo(ResetStyle, foregroundColor(b.Fg), backgroundColor(b.Bg))
				lastFg, lastBg = b.Fg, b.Bg
			}
			canvas.frame.writeTo(string(b.Rune))
			lastX, lastY = x, y
		}
	}

	//move the cursor inside the focused input box
	if p.focused != nil && p.focused.kind == Input {
		row, col := p.focused.inputState.mapCursorPos(p.focused)
		showCursorPointer()
		canvas.frame.setCursorPosition(row, col)
	}

	//move the cursor inside the focused text area box
	if p.focused != nil && p.focused.kind == TextArea {
		row, col := p.focused.inputState.mapCursorPos(p.focused)
		showCursorPointer()
		canvas.frame.setCursorPosition(row, col)
	}

	if err := canvas.frame.flush(os.Stdout); err != nil {
		return err
	}

	// front mirrors what is now on screen; keep back as the working copy so
	// partial repaints draw on top of the current Canvas.frame.
	copy(p.frontbuff.Cells, canvas.buff.Cells)

	return nil
}

func (f *Frame) writeTo(parts ...string) error {
	for _, part := range parts {
		_, err := f.wbuf.WriteString(part)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Frame) flush(w io.Writer) error {
	_, frameErr := f.wbuf.WriteTo(w)
	return frameErr
}

func (f *Frame) setCursorPosition(row, col int) {
	var numbuf [20]byte
	f.wbuf.WriteString(CSI)
	f.wbuf.Write(strconv.AppendInt(numbuf[:0], int64(row+1), 10))
	f.wbuf.WriteByte(';')
	f.wbuf.Write(strconv.AppendInt(numbuf[:0], int64(col+1), 10))
	f.wbuf.WriteByte('H')
}
