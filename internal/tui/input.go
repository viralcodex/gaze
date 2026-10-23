package tui

func (el *Element) handleInput(event KeyEvent) {
	state := &el.inputState

	switch event.KeyCode {
	case KeyLeft:
		if state.Col > 0 {
			state.Col--
		} else if state.Line > 0 {
			state.Line--
			state.Col = len(state.Text[state.Line])
		}

	case KeyRight:
		if state.Col < len(state.Text[state.Line]) {
			state.Col++
		} else if state.Line < len(state.Text)-1 {
			state.Line++
			state.Col = 0
		}

	case KeyUp:
		row, col := state.getVisualLine(el.visualLines)
		if state.DesiredX < 0 {
			state.DesiredX = col
		}
		if row > 0 {
			row = row - 1
			vl := el.visualLines[row] // previous visual line
			endOfPrevLine := vl.endCol
			if el.visualLines[row+1].Line == vl.Line {
				endOfPrevLine-- // target wraps into the row cursor was previously on, so keep it on that row
			}
			state.Line = vl.Line
			state.Col = min(vl.startCol+state.DesiredX, endOfPrevLine)
		}

	case KeyDown:
		row, col := state.getVisualLine(el.visualLines)
		if state.DesiredX < 0 {
			state.DesiredX = col
		}
		if row < len(el.visualLines)-1 {
			row = row + 1
			vl := el.visualLines[row]
			endOfNextLine := vl.endCol
			if row+1 < len(el.visualLines) && el.visualLines[row+1].Line == vl.Line {
				endOfNextLine-- // target row wraps below, so cap before the wrap column
			}
			state.Line = vl.Line
			state.Col = min(vl.startCol+state.DesiredX, endOfNextLine)
		}

	case KeyBackspace:
		if el.kind == Input && state.Xoffset > 0 {
			state.Xoffset--
		}
		if state.Col > 0 {
			line := state.Text[state.Line]
			copy(line[state.Col-1:], line[state.Col:])
			state.Text[state.Line] = line[:len(line)-1]
			state.Col--
		} else if state.Line > 0 { // merge current line with previous line
			prev := state.Text[state.Line-1]
			state.Col = len(prev)
			state.Text[state.Line-1] = append(prev, state.Text[state.Line]...)
			copy(state.Text[state.Line:], state.Text[state.Line+1:])
			state.Text = state.Text[:len(state.Text)-1]
			state.Line--
		}

	case KeyEnter:
		if el.kind != TextArea {
			break
		}
		line := state.Text[state.Line]
		tail := append([]rune{}, line[state.Col:]...)
		state.Text[state.Line] = append([]rune{}, line[:state.Col]...)
		state.Text = append(state.Text[:state.Line+1], append([][]rune{tail}, state.Text[state.Line+1:]...)...)
		state.Line++
		state.Col = 0

	case KeyRune:
		if event.Mods.Ctrl || event.Mods.Alt || event.Rune < 0x20 {
			return
		}
		if el.MaxLength > 0 && len(state.Text[state.Line]) >= el.MaxLength {
			return
		}
		line := append(state.Text[state.Line], 0)
		copy(line[state.Col+1:], line[state.Col:])
		line[state.Col] = event.Rune
		state.Text[state.Line] = line
		state.Col++
	}

	// One place keeps wrapping, the sticky column and scrolling in sync.
	if el.kind == TextArea {
		setVisualLines(el)
		row, col := state.getVisualLine(el.visualLines)
		if event.KeyCode != KeyUp && event.KeyCode != KeyDown {
			state.DesiredX = col
		}
		state.clampCursorY(row, el.contentRect.H)
	} else {
		state.clampCursorX(state.Col, el.contentRect.W)
	}
}

func (state *InputState) clampCursorX(col, width int) {
	if width <= 0 {
		return
	}

	if col < state.Xoffset {
		state.Xoffset = col
	}
	if col > state.Xoffset+width-1 {
		state.Xoffset = col - width + 1
	}
}

func (state *InputState) clampCursorY(row, height int) {
	if height <= 0 {
		return
	}

	if row < state.Yoffset {
		state.Yoffset = row
	}
	if row > state.Yoffset+height-1 {
		state.Yoffset = row - height + 1
	}
}

// func (state *InputState) backspace(el *Element) {
// 	copy(state.Text[state.Line][state.Col-1:], state.Text[state.Line][state.Col:])
// 	state.Text[state.Line] = state.Text[state.Line][:len(state.Text[state.Line])-1]
// 	state.Col--
// 	if el.kind == Input {
// 		state.clampCursor(el)
// 	}
// }
