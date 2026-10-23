package tui

func (p *Program) handleKeyEvent(event KeyEvent) bool {
	if event.Mods.Ctrl && event.Rune == 'c' {
		return true
	}

	focusedInput := p.focused

	if focusedInput == nil || (focusedInput.kind != Input && focusedInput.kind != TextArea) {
		return false
	}

	switch focusedInput.kind {
	case Input, TextArea:
		focusedInput.handleInput(event)
		if focusedInput.OnInput != nil {
			focusedInput.OnInput(focusedInput, event)
		}
		p.MarkDirty(focusedInput)
		return false
	}

	return false
}
