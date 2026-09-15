package tui

import (
	"fmt"
	"slices"
)

func parseMouseEvent(input string) (Event, bool) {
	var mouseEvent MouseEvent
	var keyPress rune

	n, err := fmt.Sscanf(input, MouseEventScanFormat, &mouseEvent.Button, &mouseEvent.X, &mouseEvent.Y, &keyPress)

	if err != nil || n != 4 || (keyPress != 'M' && keyPress != 'm') {
		return Event{}, false
	}

	switch {
	case keyPress == 'm':
		mouseEvent.Action = MouseRelease
	case mouseEvent.Button&32 != 0:
		mouseEvent.Action = MouseMove
	default:
		mouseEvent.Action = MousePress
	}

	event := Event{
		Kind:       Mouse,
		MouseEvent: mouseEvent,
	}

	return event, true
}

func parseKeyEvent(input []byte) (Event, int) {
	return Event{
		Kind: Key,
		KeyEvent: KeyEvent{
			Buffer: input,
			N:      len(input),
		},
	}, len(input)
}

func (p *Program) handleKeyEvent(event KeyEvent) (dirty *Element, quit bool) {
	if event.N == 0 {
		return nil, false
	}

	if event.Buffer[0] == 3 {
		return nil, true
	}

	// fmt.Printf("%q\n", string(event.Buffer))

	focusedInput := p.focused

	if focusedInput == nil || focusedInput.kind != Input {
		return nil, false
	}

	switch focusedInput.kind {
	case Input:
		focusedInput.handleInput(event)
		if focusedInput.OnInput != nil {
			focusedInput.OnInput(focusedInput, event)
		}
		return focusedInput, false
	}

	return nil, false
}

func (el *Element) handleInput(event KeyEvent) {

	state := &el.inputState

	switch event.KeyCode {
	case KeyLeft:
		if state.Cursor > 0 {
			state.Cursor--
		}
		if state.ScrollOffset > 0 {
			state.ScrollOffset--
		}
	case KeyRight:
		if state.Cursor < len(state.Text) {
			state.Cursor++
			if state.Cursor >= el.contentRect.W {
				state.ScrollOffset++
			}
		}
	case KeyRune:
		if event.Rune == 0x7F || event.Rune == 0x08 {
			if state.Cursor <= 0 {
				return
			}
			if state.ScrollOffset > 0 {
				state.ScrollOffset--
			}
			copy(state.Text[state.Cursor-1:], state.Text[state.Cursor:])
			state.Text = state.Text[:len(state.Text)-1]
			state.Cursor--
		} else {
			if el.MaxLength > 0 && len(state.Text) >= el.MaxLength {
				return
			}
			state.Text = append(state.Text, 0)
			copy(state.Text[state.Cursor+1:], state.Text[state.Cursor:])
			state.Text[state.Cursor] = event.Rune
			state.Cursor++
			state.clampCursorAtEnd(el.contentRect.W)
		}
	}
}

func (state *InputState) clampCursorAtEnd(width int) {
	if width <= 0 {
		return
	}

	if state.Cursor < state.ScrollOffset {
		state.ScrollOffset = state.Cursor
	}
	if state.Cursor > state.ScrollOffset+width-1 {
		state.ScrollOffset = state.Cursor - width + 1
	}
}

func (p *Program) handleMouseEvent(event MouseEvent) []*Element {
	var dirtyElements []*Element

	markDirty := func(el *Element) {
		if el == nil {
			return
		}
		if slices.Contains(dirtyElements, el) {
			return
		}
		dirtyElements = append(dirtyElements, el)
	}

	current := p.hitMap.at(event.X-1, event.Y-1) //0-based hitmap
	previous := p.mouseState.Hovered

	//hover
	if current != previous {
		if previous != nil {
			previous.state.Hovered = false
			if previous.Style.Hover != nil {
				markDirty(previous)
			}
			if previous.OnMouseOut != nil {
				previous.OnMouseOut(previous, event)
			}
		}

		p.mouseState.Hovered = current

		if current != nil {
			current.state.Hovered = true
			if current.Style.Hover != nil {
				markDirty(current)
			}

			if current.OnMouseEnter != nil {
				current.OnMouseEnter(current, event)
			}
			p.applyCursor(current.Style.Cursor)
		} else {
			p.applyCursor(CursorDefault)
		}
	}

	//click
	switch event.Action {
	case MousePress:
		p.mouseState.Pressed = current
		if current != nil {
			current.state.Pressed = true
			markDirty(current)
		}
	case MouseRelease:
		pressed := p.mouseState.Pressed
		if pressed != nil {
			pressed.state.Pressed = false
			pressed.state.Focused = false
			p.focused = nil
			markDirty(pressed)
		}

		p.mouseState.Pressed = nil
		p.focused = nil

		if current != nil && current == pressed {
			if current.OnClick != nil {
				current.OnClick(current, event)
			}
			current.state.Focused = true
			p.focused = current
		}
	}

	return dirtyElements
}

// applyCursor changes the terminal cursor shape only when it differs from the
// currently applied one, so repeated moves over the same element write nothing.
func (p *Program) applyCursor(c Cursor) {
	if c == p.cursor {
		return
	}
	p.cursor = c
	setCursorShape(c)
}

func (p *Program) debugSnapshot(event Event, hit, clicked *Element) DebugSnapshot {
	return DebugSnapshot{
		Parsed:   fmt.Sprintf("%q", event.raw),
		Buffered: fmt.Sprintf("%q", event.buffered),
		Mouse:    event.MouseEvent,
		Hit:      elementID(hit),
		Hovered:  elementID(p.mouseState.Hovered),
		Pressed:  elementID(p.mouseState.Pressed),
		Clicked:  elementID(clicked),
		Focused:  elementID(p.focused),
	}
}

func elementID(el *Element) string {
	if el == nil {
		return "-"
	}
	return el.ID
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
