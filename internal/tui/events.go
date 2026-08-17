package tui

import (
	"fmt"
)

type HitMap struct {
	Width    int
	Height   int
	Cells    []uint32
	Elements []*Element
}

type MouseState struct {
	Hovered *Element
	Pressed *Element
}

type MouseAction uint8

const (
	MouseMove MouseAction = iota
	MousePress
	MouseRelease
)

func ParseMouseEvent(input string) (MouseEvent, bool) {
	var event MouseEvent
	var keyPress rune

	n, err := fmt.Sscanf(input, MouseEventScanFormat, &event.Button, &event.X, &event.Y, &keyPress)

	if err != nil || n != 4 || (keyPress != 'M' && keyPress != 'm') {
		return MouseEvent{}, false
	}

	switch {
	case keyPress == 'm':
		event.Action = MouseRelease
	case event.Button&32 != 0:
		event.Action = MouseMove
	default:
		event.Action = MousePress
	}

	return event, true
}

func HandleKeyEvent(event KeyEvent) bool {
	if len(event.Buffer) > 0 && event.N > 0 {
		byt := event.Buffer[0]
		if byt == 3 {
			return true
		}
		// if byt == 127 && len(terminalState.text) > 0 {
		// terminalState.text = terminalState.text[:len(terminalState.text)-1]
		// } else if byt >= 32 && byt <= 126 {
		// terminalState.text = append(terminalState.text, byt)
		// }
	}
	return false
}

func HandleMouseEvent(event MouseEvent) bool {
	isDirty := false
	current := hitMap.At(event.X, event.Y)
	previous := mouseState.Hovered

	//hover
	if current != previous {
		if previous != nil {
			previous.State.Hovered = false
			if previous.OnMouseOut != nil {
				previous.OnMouseOut(previous, event)
			}
		}

		mouseState.Hovered = current

		if current != nil {
			current.State.Hovered = true
			if current.OnMouseEnter != nil {
				current.OnMouseEnter(current, event)
			}
		}
		isDirty = true
	}
	//click
	switch event.Action {
	case MousePress:
		mouseState.Pressed = current
		if current != nil {
			current.State.Pressed = true
			isDirty = true
		}
	case MouseRelease:
		pressed := mouseState.Pressed
		if pressed != nil {
			pressed.State.Pressed = false
			isDirty = true
		}
		mouseState.Pressed = nil

		if current != nil && current == pressed && current.OnClick != nil {
			current.OnClick(current, event)
		}

	}

	return isDirty
}

// func hitCheck(rect Rect, x, y int) bool {
// 	return x >= rect.X &&
// 		x < rect.X+rect.W &&
// 		y >= rect.Y &&
// 		y < rect.Y+rect.H
// }

func (m *HitMap) At(x, y int) *Element {
	x--
	y--

	if x < 0 || y < 0 || x >= m.Width || y >= m.Height {
		return nil
	}

	id := m.Cells[y*m.Width+x]

	if id == 0 {
		return nil
	}

	return m.Elements[id]
}
