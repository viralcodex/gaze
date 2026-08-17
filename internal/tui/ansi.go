package tui

import "fmt"

const (
	esc = "\x1b"
	bel = "\x07"

	CSI = esc + "["
	OSC = esc + "]"

	AltScreenEnter       = CSI + "?1049h"
	AltScreenExit        = CSI + "?1049l"
	ClearScreen          = CSI + "2J"
	CursorHome           = CSI + "H"
	ClearLine            = CSI + "2K"
	KittyGraphicsStart   = esc + "_G"
	KittyGraphicsEnd     = esc + "\\"
	MouseClickEnable     = CSI + "?1000h"
	MouseClickDisable    = CSI + "?1000l"
	MouseMotionEnable    = CSI + "?1003h"
	MouseMotionDisable   = CSI + "?1003l"
	MouseSGREnable       = CSI + "?1006h"
	MouseSGRDisable      = CSI + "?1006l"
	MouseEventScanFormat = CSI + "<%d;%d;%d%c"
	CursorPosition       = CSI + "%d;%dH"
	PointerCursor        = OSC + "22;pointer" + bel
	DefaultCursor        = OSC + "22;" + bel
	ShowCursor           = CSI + "?25h"
	HideCursor           = CSI + "?25l"
	ForegroundColor      = CSI + "38;2;%d;%d;%dm"
	BackgroundColor      = CSI + "48;2;%d;%d;%dm"
	ResetBackground      = CSI + "49m"
	ResetStyle           = CSI + "0m"
)

func foregroundColor(color Color) string {
	if !color.Set {
		return ""
	}
	return fmt.Sprintf(ForegroundColor, color.R, color.G, color.B)
}

func backgroundColor(color Color) string {
	if !color.Set {
		return ""
	}
	return fmt.Sprintf(BackgroundColor, color.R, color.G, color.B)
}
