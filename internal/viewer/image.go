package viewer

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"gaze/internal/tui"
)

// the image data doesn't need to be sent again until:
// it changes internally (cropped, rotate, filter, etc)
// terminal state changes (restart, UI changes, etc)

type Image struct {
	Id         uint32
	FileName   string
	Path       string
	Data       []byte
	Dirty      bool
	Uploaded   bool
	Dimensions ImageDimensions
	Rect       ImageRect
	State      ImageState
}

type ImageState struct {
	Rotation int
	Zoom     float32
}

type ImageDimensions struct {
	Width  int
	Height int
}

type ImageRect struct {
	Cols int
	Rows int
}

var nextImageID uint32

var allowedImgTypes = []string{".jpg", ".jpeg", ".png", ".webp"}

func (img Image) getImgSource() tui.ImageSource {
	return tui.ImageSource{
		ID:          img.Id,
		Data:        img.Data,
		NeedsUpload: true,
		Dimensions: tui.ImageDimensions{
			Width:  img.Dimensions.Width,
			Height: img.Dimensions.Height,
		},
	}
}

func LoadImage(path string) (Image, error) {
	fileName := filepath.Base(path)

	if !verifyFileType(path) {
		return Image{}, fmt.Errorf("unsupported image type for file: %s", fileName)
	}

	imgData, err := os.ReadFile(path)

	if err != nil {
		return Image{}, fmt.Errorf("read image %q: %w", path, err)
	}

	imgReader := bytes.NewReader(imgData)

	config, format, err := image.DecodeConfig(imgReader)

	if err != nil {
		return Image{}, fmt.Errorf("error decoding image config: %s", fileName)
	}

	if format != "png" {
		imgData, err = convertToPng(imgReader, fileName)
		if err != nil {
			return Image{}, err
		}
	}
	return Image{
		Id:       newImageID(),
		FileName: fileName,
		Path:     path,
		Data:     imgData,
		Dimensions: ImageDimensions{
			Width:  config.Width,
			Height: config.Height,
		},
	}, nil
}

func convertToPng(imgReader *bytes.Reader, fileName string) ([]byte, error) {
	imgReader.Seek(0, 0) //start reading from the start
	decodedImgData, _, err := image.Decode(imgReader)

	if err != nil {
		return nil, fmt.Errorf("error decoding image: %s", fileName)
	}

	var pngBytes bytes.Buffer

	err = png.Encode(&pngBytes, decodedImgData)

	if err != nil {
		return nil, fmt.Errorf("error encoding to png image: %s", fileName)
	}

	return pngBytes.Bytes(), nil
}

func verifyFileType(path string) bool {
	ext := filepath.Ext(strings.ToLower(path))
	for _, extension := range allowedImgTypes {
		if ext == strings.ToLower(extension) {
			return true
		}
	}
	return false
}

func sendImageData() {
	tui.SetImageData(img.getImgSource())
}

func getImageRect() {
	cols, rows := tui.FitToRect(tui.Rect{
		Y: 4,
		W: terminalState.Dimensions.Width,
		H: terminalState.Dimensions.Height - 5,
	})
	img.Rect = ImageRect{
		Cols: cols,
		Rows: rows,
	}
}

// these ops send the updated image data to tui (rendered = false)
func zoomImage(el *tui.Element) {
}

func rotateImage(el *tui.Element) {

}

func newImageID() uint32 {
	nextImageID++
	return nextImageID
}
