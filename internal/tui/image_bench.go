package tui

import (
	"testing"
)

//MOVE THIS BACK TO TUI MODULE AND RUN WITH COMMAND BELOW
//go test ./internal/tui/ -run '^$' -bench 'BenchmarkUploadImage' -benchmem -count=1 2>&1 | tail -15

// newBenchCanvas builds a Canvas whose Frame writes into an in-memory buffer,
// so we measure encode + buffer cost only, not terminal I/O.
func newBenchCanvas() *Canvas {
	return &Canvas{
		frame:     &Frame{},
		imageData: map[uint32]*ImageSource{},
	}
}

func makeImage(id uint32, sizeBytes int) *ImageSource {
	data := make([]byte, sizeBytes)
	for i := range data {
		data[i] = byte(i)
	}
	return &ImageSource{ID: id, Data: data, NeedsUpload: true}
}

var imageSizes = []struct {
	name  string
	bytes int
}{
	{"64KB", 64 << 10},
	{"512KB", 512 << 10},
	{"2MB", 2 << 20},
	{"8MB", 8 << 20},
}

// BenchmarkUploadImage measures a full re-transmit (kitty a=t): base64 encode
// the whole payload plus chunked writes. This is what resize() triggers today
// via markImageReupload.
func BenchmarkUploadImage(b *testing.B) {
	for _, sz := range imageSizes {
		b.Run(sz.name, func(b *testing.B) {
			c := newBenchCanvas()
			dc := &DrawContext{Canvas: c}
			img := makeImage(1, sz.bytes)
			b.ReportAllocs()
			b.SetBytes(int64(sz.bytes))
			for b.Loop() {
				c.frame.wbuf.Reset()
				img.NeedsUpload = true
				if err := dc.uploadImage(img); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkPlaceImage measures a re-placement (kitty a=p): a single control
// sequence, no pixel data. This is what a resize should trigger when only the
// geometry changed.
func BenchmarkPlaceImage(b *testing.B) {
	for _, sz := range imageSizes {
		b.Run(sz.name, func(b *testing.B) {
			c := newBenchCanvas()
			dc := &DrawContext{Canvas: c}
			img := makeImage(1, sz.bytes)
			b.ReportAllocs()
			for b.Loop() {
				c.frame.wbuf.Reset()
				if err := dc.placeImage(img, 0, 0, 80, 24); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
