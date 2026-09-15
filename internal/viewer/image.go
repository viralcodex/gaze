package viewer

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"gaze/internal/tui"
)

// the image data doesn't need to be sent again until:
// it changes internally (cropped, rotate, filter, etc)
// terminal state changes (restart, UI changes, etc)

type ImageData struct {
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

func loadImage(path string) (ImageData, error) {
	fileName := filepath.Base(path)

	if !verifyFileType(path) {
		return ImageData{}, fmt.Errorf("unsupported image type for file: %s", fileName)
	}

	imgData, err := os.ReadFile(path)

	if err != nil {
		return ImageData{}, fmt.Errorf("read image %q: %w", path, err)
	}

	imgReader := bytes.NewReader(imgData)

	config, format, err := image.DecodeConfig(imgReader)

	if err != nil {
		return ImageData{}, fmt.Errorf("error decoding image config: %s:%v", fileName, err)
	}

	if format != "png" {
		imgData, err = convertToPng(imgReader, fileName)
		if err != nil {
			return ImageData{}, err
		}
	}
	return ImageData{
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

func registerImage() {
	program.SetImageData(terminalState.ImageData.getImgSource())
}

func (img ImageData) getImgSource() tui.ImageSource {
	return tui.ImageSource{
		ID:          img.Id,
		Data:        img.Data,
		NeedsUpload: true,
	}
}

// these ops send the updated image data to tui (rendered = false)
func zoomImage(el *tui.Element) {
	el.Label = "--clicked--"
}

func rotateImage(el *tui.Element) {
	el.Label = "--clicked--"
}

func newImageID() uint32 {
	nextImageID++
	return nextImageID
}
