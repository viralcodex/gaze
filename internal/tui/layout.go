package tui

func Draw(el *Element) {
	if el == nil || !el.State.Visible {
		return
	}

	switch el.Kind {
	case ElementBox:
		// drawBox(el)
	case ElementButton:
		drawButton(el)
	case ElementInput:
		// drawInput(el)
	case ElementImage:
		drawImage(el)
	}
	for _, child := range el.Children {
		Draw(child)
	}
}
