package tui

import (
	"io"
	"os"
	"strconv"
)

func (b *Buffer) fill(rect Rect, r rune, cell Cell) {
	for y := rect.Y; y < rect.Y+rect.H; y++ {
		for x := rect.X; x < rect.X+rect.W; x++ {
			b.set(x, y, r, cell)
		}
	}
}

func (b *Buffer) set(x, y int, r rune, cell Cell) {
	if x < 0 || y < 0 || x >= b.Width || y >= b.Height {
		return
	}
	cell.Rune = r
	b.Cells[y*b.Width+x] = cell
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
				canvas.frame.cursorPosition(y, x)
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
		row := p.focused.contentRect.Y + (p.focused.contentRect.H / 2)
		col := p.focused.contentRect.X + p.focused.inputState.Cursor - p.focused.inputState.ScrollOffset
		canvas.frame.cursorPosition(row, col)
	}

	if err := canvas.frame.flush(os.Stdout); err != nil {
		return err
	}

	// front mirrors what is now on screen; keep back as the working copy so
	// partial repaints draw on top of the current Canvas.frame, not a stale one.
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
	// if _, err := f.wbuf.WriteString(BeginSynchronizedUpdate); err != nil {
	// 	return err
	// }

	_, frameErr := f.wbuf.WriteTo(w)
	// _, endErr := f.wbuf.WriteString(EndSynchronizedUpdate)
	// if frameErr != nil {
	// 	return frameErr
	// }
	return frameErr
}

func (f *Frame) cursorPosition(row, col int) {
	var numbuf [20]byte
	f.wbuf.WriteString(CSI)
	f.wbuf.Write(strconv.AppendInt(numbuf[:0], int64(row+1), 10))
	f.wbuf.WriteByte(';')
	f.wbuf.Write(strconv.AppendInt(numbuf[:0], int64(col+1), 10))
	f.wbuf.WriteByte('H')
}
