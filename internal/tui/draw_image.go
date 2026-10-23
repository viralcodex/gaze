package tui

import "fmt"

func drawImage(el *Element, dc *DrawContext) error {
	outer := el.computedRect
	contentRect := el.contentRect
	style := styleFor(el)

	cols, rows := dc.fitToRect(el)

	x := contentRect.X + alignOffset(el.ImageRef.Position.X, contentRect.W-cols)
	y := contentRect.Y + alignOffset(el.ImageRef.Position.Y, contentRect.H-rows)

	currentImage, ok := dc.Canvas.imageData[el.ImageRef.ID]

	if !ok {
		return fmt.Errorf("%s", "No image found")
	}

	if currentImage.NeedsUpload {
		if err := dc.uploadImage(currentImage); err != nil {
			return err
		}
	}

	if dc.Canvas.lastPlacedImageID != 0 && dc.Canvas.lastPlacedImageID != currentImage.ID {
		dc.deleteImagePlacement(dc.Canvas.lastPlacedImageID)
	}

	dc.fill(outer, ' ', Cell{Bg: style.bgColor})

	if err := dc.placeImage(currentImage, x, y, cols, rows); err != nil {
		return err
	}
	dc.Canvas.lastPlacedImageID = currentImage.ID
	return nil
}

// alignOffset places a fitted span of length `free` (box - image) along one
// axis. Shared with the flex leading-offset logic in justifyOffsets.
func alignOffset(a Align, free int) int {
	if free <= 0 {
		return 0
	}
	switch a {
	case AlignCenter:
		return free / 2
	case AlignEnd:
		return free
	}
	return 0 // AlignStart
}
