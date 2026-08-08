package tui

import (
	"fmt"
)

func ParseMouseEvent(input string) (MouseEvent, bool) {
	var event MouseEvent
	var keyPress rune

	n, err := fmt.Sscanf(input, MouseEventScanFormat, &event.Button, &event.X, &event.Y, &keyPress)

	if err != nil || n != 4 || (keyPress != 'M' && keyPress != 'm') {
		return MouseEvent{}, false
	}

	motion := event.Button&32 != 0
	noButton := event.Button&3 == 3

	event.Press = keyPress == 'M' && !motion && !noButton

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
	fmt.Print(cursorPosition(5, 1)) // row 3, column 1
	fmt.Print(ClearLine)            // clear current line
	fmt.Printf("Mouse data: key=%d x=%d y=%d pressed=%v\n", event.Button, event.X, event.Y, event.Press)
	// isInsideButton := hitCheck(event.X, event.Y)

	// if event.Press {
	// 	terminalState.button.pressed = isInsideButton
	// 	return false
	// }

	// if terminalState.button.pressed {
	// 	terminalState.button.pressed = false
	// 	if isInsideButton && terminalState.button.onClick != nil {
	// 		terminalState.button.onClick()
	// 	}
	// }

	// if isInsideButton {
	// 	fmt.Print(pointerCursor)
	// } else {
	// 	fmt.Print(defaultCursor)
	// }

	// fmt.Print(cursorPosition(4, 1)) // row 3, column 1
	// fmt.Print(clearLine)            // clear current line
	// fmt.Printf("Button hovered: %v", isInsideButton)

	// return isInsideButton
	return true
}

func hitCheck(rect Rect, x, y int) bool {
	return x >= rect.X &&
		x < rect.X+rect.W &&
		y >= rect.Y &&
		y < rect.Y+rect.H
}