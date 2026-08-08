package tui

import (
	"encoding/base64"
	"fmt"
	"math"
	"os"
)

type ImageSource struct {
	ID          uint32
	Data        []byte
	NeedsUpload bool
	Dimensions  ImageDimensions
}

type ImageDimensions struct {
	Width  int
	Height int
}

const chunkSize = 4096

var currentImage ImageSource

func SetImageData(image ImageSource) {
	currentImage = image
}

func MarkImageReupload() {
	currentImage.NeedsUpload = true
}

func UploadImageData(cols, rows int) error {
	encodedImageData := base64.StdEncoding.EncodeToString(currentImage.Data)

	for offset := 0; offset < len(encodedImageData); offset += chunkSize {
		end := min(offset+chunkSize, len(encodedImageData))

		hasMore := end < len(encodedImageData)
		hasMoreFlag := 0

		chunk := encodedImageData[offset:end]

		if hasMore {
			hasMoreFlag = 1
		}

		controlBytes := fmt.Sprintf("a=T,i=%d,f=100,t=d,c=%d,r=%d,m=%d", currentImage.ID, cols, rows, hasMoreFlag)

		if offset > 0 {
			controlBytes = fmt.Sprintf("m=%d", hasMoreFlag)
		}

		if _, err := fmt.Fprintf(os.Stdout, "%s%s;%s%s", KittyGraphicsStart, controlBytes, chunk, KittyGraphicsEnd); err != nil {
			return err
		}
	}
	currentImage.NeedsUpload = false
	return nil
}

func PlaceImage(x, y, cols, rows int) error {
	fmt.Print(cursorPosition(y, x))

	controlBytes := fmt.Sprintf("a=p,i=%d,c=%d,r=%d", currentImage.ID, cols, rows)
	_, err := fmt.Fprintf(os.Stdout, "%s%s;%s", KittyGraphicsStart, controlBytes, KittyGraphicsEnd)
	return err
}

func FitToRect(rect Rect) (int, int) {
	imgW := currentImage.Dimensions.Width
	imgH := currentImage.Dimensions.Height

	maxCols := rect.W
	maxRows := rect.H

	if imgW <= 0 || imgH <= 0 || maxCols <= 0 || maxRows <= 0 {
		return 1, 1
	}

	const terminalAspectRatio = 1.75

	imgAspectRatio := float64(imgW) / float64(imgH)
	cellAspect := imgAspectRatio * terminalAspectRatio

	cols := maxCols
	rows := int(math.Round(float64(maxCols) / cellAspect))

	if rows > maxRows {
		rows = maxRows
		cols = int(math.Round(float64(rows) * cellAspect))
	}

	cols = int(math.Max(1, float64(cols)))
	rows = int(math.Max(1, float64(rows)))

	return cols, rows
}
