package tui

import (
	"bytes"
	"io"
)

type Frame struct {
	wbuf bytes.Buffer
}

func (f *Frame) writeOut(parts ...string) error {
	for _, part := range parts {
		_, err := f.wbuf.WriteString(part)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Frame) flush(w io.Writer) error {
	_, err := f.wbuf.WriteTo(w)
	return err
}

func (f *Frame) reset() {
	frame.wbuf.Reset()
}
