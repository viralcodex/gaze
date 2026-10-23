package tui

// mouse button stores raw byte value, which is then interpreted to determine the action (press, release, scroll, move)
// bits	value	used for
// 0-1	0/1/2/3 left, middle, right, none (buttons)
// 2	4		Shift
// 3	8		Alt/Meta
// 4	16		Ctrl
// 5	32		motion (drag/move)
// 6	64		wheel
// 7	128		extra buttons (8-11)
// thus we have to normalise the raw byte value after determining the action to get the actual button pressed
func parseMouseEvent(inputBytes []byte, semiColonIndices []int) (Event, bool) {
	var mouseEvent MouseEvent
	var pressed rune

	setMouseEvent(inputBytes, &mouseEvent, &pressed, semiColonIndices)

	raw := mouseEvent.Button

	switch {
	case pressed == 'm':
		mouseEvent.Action = MouseRelease
	// Scroll events: bit 64 set. Low bits select direction (0 = up, 1 = down).
	case raw&64 != 0:
		if raw&1 != 0 {
			mouseEvent.Action = MouseScrollDown
		} else {
			mouseEvent.Action = MouseScrollUp
		}
	case raw&32 != 0:
		mouseEvent.Action = MouseMove
	default:
		mouseEvent.Action = MousePress
	}

	//now normalise the raw bytes for button plus modifier keys
	decodeButtonAndMods(raw, &mouseEvent) //mask and decode modifier keys

	event := Event{
		Kind:       Mouse,
		MouseEvent: mouseEvent,
	}

	return event, true
}

func setMouseEvent(inputBytes []byte, mouseEvent *MouseEvent, pressed *rune, semiColonIndices []int) {
	*pressed = rune(inputBytes[len(inputBytes)-1])

	button := inputBytes[3:semiColonIndices[0]]
	x := inputBytes[semiColonIndices[0]+1 : semiColonIndices[1]]
	y := inputBytes[semiColonIndices[1]+1 : len(inputBytes)-1]

	mouseEvent.Button = bytesToInt(button) //raw bytes stored, normalise after action
	mouseEvent.X = bytesToInt(x)
	mouseEvent.Y = bytesToInt(y)
}

func bytesToInt(b []byte) int {
	n := 0
	for _, c := range b {
		n = n*10 + int(c-'0')
	}
	return n
}

func decodeButtonAndMods(raw int, mouseEvent *MouseEvent) {
	mouseEvent.Button = (raw & 3) //mask everything but the button (0-1 bits)

	if raw&128 != 0 {
		mouseEvent.Button += 8 //extra buttons (8-11)
	}

	shift := raw&4 != 0
	alt := raw&8 != 0
	ctrl := raw&16 != 0

	//wheel events have no button, scroll is already recorded in the MouseAction
	if raw&64 != 0 {
		mouseEvent.Button = -1
	}

	mouseEvent.Mods = Mods{
		Shift: shift,
		Alt:   alt,
		Ctrl:  ctrl,
	}
}

func (p *Program) handleMouseEvent(event MouseEvent) {
	var dirtyElements []*Element

	current := p.hitMap.at(event.X-1, event.Y-1) //0-based hitmap
	previous := p.mouseState.Hovered

	//hover
	if current != previous {
		if previous != nil {
			previous.state.Hovered = false
			if previous.Style.Hover != nil {
				dirtyElements = append(dirtyElements, previous)
			}
			if previous.OnMouseOut != nil {
				previous.OnMouseOut(previous, event)
			}
		}

		p.mouseState.Hovered = current

		if current != nil {
			current.state.Hovered = true
			if current.Style.Hover != nil {
				dirtyElements = append(dirtyElements, current)
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
			dirtyElements = append(dirtyElements, current)

		}
	case MouseRelease:
		pressed := p.mouseState.Pressed
		if pressed != nil {
			pressed.state.Pressed = false
			pressed.state.Focused = false
			p.focused = nil
			dirtyElements = append(dirtyElements, pressed)
		}

		p.mouseState.Pressed = nil
		p.focused = nil
		hideCursorPointer()

		if current != nil && current == pressed {
			if current.OnClick != nil {
				current.OnClick(current, event)
			}
			if current.Focusable {
				current.state.Focused = true
				p.focused = current
			}
		}
	}

	p.MarkDirty(dirtyElements...)
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
