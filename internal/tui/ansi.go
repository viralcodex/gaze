package tui

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
	PointerCursor        = OSC + "22;pointer" + bel
	DefaultCursor        = OSC + "22;" + bel
)