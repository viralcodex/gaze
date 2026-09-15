package tui

import (
	"encoding/base64"
	"math"
	"strconv"
)

const chunkSize = 4096
const rawChunk = chunkSize / 4 * 3 // a 4,096-character chunk of Base64 text will decode into exactly 3,072 bytes of original binary data
const imagePlacementID = 1

func (p *Program) SetImageData(image ImageSource) {
	p.Canvas.imageData[image.ID] = &image
}

func (c *Canvas) markImageReupload() {
	for _, img := range c.imageData {
		img.NeedsUpload = true
	}
}

func (c *Canvas) uploadImage(image *ImageSource) error {
	var buf [chunkSize]byte
	var numbuf [20]byte
	w := &c.frame.wbuf

	for offset := 0; offset < len(image.Data); offset += rawChunk {

		end := min(offset+rawChunk, len(image.Data))

		rawBytes := image.Data[offset:end]
		encodedChunks := buf[:base64.StdEncoding.EncodedLen(len(rawBytes))]

		base64.StdEncoding.Encode(encodedChunks, rawBytes) //encode rawBytes into chunk

		hasMore := end < len(image.Data)

		w.WriteString(KittyGraphicsStart)
		if offset == 0 {
			w.WriteString("a=t,i=")
			w.Write(strconv.AppendUint(numbuf[:0], uint64(image.ID), 10))
			w.WriteString(",f=100,t=d,m=")
		} else {
			w.WriteString("m=")
		}

		if hasMore {
			w.WriteByte('1')
		} else {
			w.WriteByte('0')
		}

		w.WriteString(",q=2;")
		w.Write(encodedChunks)
		w.WriteString(KittyGraphicsEnd)
	}

	image.NeedsUpload = false

	return nil
}

func (c *Canvas) placeImage(image *ImageSource, x, y, cols, rows int) error {
	var numbuf [20]byte
	f := c.frame
	w := &f.wbuf

	f.cursorPosition(y, x)

	w.WriteString(KittyGraphicsStart)
	w.WriteString("a=p,i=")
	w.Write(strconv.AppendUint(numbuf[:0], uint64(image.ID), 10))
	w.WriteString(",p=")
	w.Write(strconv.AppendUint(numbuf[:0], uint64(imagePlacementID), 10))
	w.WriteString(",c=")
	w.Write(strconv.AppendUint(numbuf[:0], uint64(cols), 10))
	w.WriteString(",r=")
	w.Write(strconv.AppendUint(numbuf[:0], uint64(rows), 10))
	w.WriteString(",q=2;")
	w.WriteString(KittyGraphicsEnd)

	return nil
}

func (c *Canvas) fitToRect(el *Element) (int, int) {
	imgW := el.ImageRef.Dimensions.Width
	imgH := el.ImageRef.Dimensions.Height

	maxCols := el.contentRect.W
	maxRows := el.contentRect.H

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
