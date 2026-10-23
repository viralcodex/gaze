package tui

import (
	"reflect"
	"testing"
)

func TestNewBoxKindAndDefaults(t *testing.T) {
	box := NewBox(BoxProps{ID: "box"})

	if box.kind != Box {
		t.Fatalf("kind = %q, want %q", box.kind, Box)
	}
	if box.Draw == nil {
		t.Fatal("Draw is nil")
	}
	if !box.state.Visible {
		t.Fatal("box not visible by default")
	}
}

func TestNewBoxPreservesChildOrder(t *testing.T) {
	a := NewText(TextProps{ID: "a"})
	b := NewText(TextProps{ID: "b"})
	c := NewText(TextProps{ID: "c"})

	box := NewBox(BoxProps{ID: "box"}, a, b, c)

	got := []string{box.children[0].ID, box.children[1].ID, box.children[2].ID}
	want := []string{"a", "b", "c"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("child order = %v, want %v", got, want)
	}
}

func TestNewButtonFieldsAndHandlers(t *testing.T) {
	clicked := false
	btn := NewButton(ButtonProps{
		ID:      "btn",
		Label:   "click me",
		OnClick: func(*Element, MouseEvent) { clicked = true },
	})

	if btn.kind != Button {
		t.Fatalf("kind = %q, want %q", btn.kind, Button)
	}
	if btn.Label != "click me" {
		t.Fatalf("Label = %q, want %q", btn.Label, "click me")
	}
	if btn.OnMouseEnter != nil || btn.OnMouseOut != nil {
		t.Fatal("nil handlers should stay nil")
	}
	if btn.OnClick == nil {
		t.Fatal("OnClick is nil")
	}
	btn.OnClick(btn, MouseEvent{})
	if !clicked {
		t.Fatal("OnClick not wired to element")
	}
}

func TestNewInputFields(t *testing.T) {
	in := NewInput(InputProps{
		ID:          "in",
		Placeholder: "type",
		MaxLength:   5,
		AutoFocus:   true,
	})

	if in.kind != Input {
		t.Fatalf("kind = %q, want %q", in.kind, Input)
	}
	if in.Placeholder != "type" {
		t.Fatalf("Placeholder = %q, want %q", in.Placeholder, "type")
	}
	if in.MaxLength != 5 {
		t.Fatalf("MaxLength = %d, want 5", in.MaxLength)
	}
	if !in.autoFocus {
		t.Fatal("autoFocus = false, want true")
	}
	if in.inputState.Text == nil {
		t.Fatal("inputState.Text not initialized")
	}
}

func TestNewTextMapsTextToLabel(t *testing.T) {
	txt := NewText(TextProps{ID: "t", Text: "hello"})

	if txt.kind != Text {
		t.Fatalf("kind = %q, want %q", txt.kind, Text)
	}
	if txt.Label != "hello" {
		t.Fatalf("Label = %q, want %q", txt.Label, "hello")
	}
}

func TestNewImageKeepsRef(t *testing.T) {
	ref := ImageRef{ID: 42}
	img := NewImage(ImageProps{ID: "img", ImageRef: ref})

	if img.kind != Image {
		t.Fatalf("kind = %q, want %q", img.kind, Image)
	}
	if img.ImageRef.ID != 42 {
		t.Fatalf("ImageRef.ID = %d, want 42", img.ImageRef.ID)
	}
}

func TestStyleColorsResolved(t *testing.T) {
	box := NewBox(BoxProps{
		ID: "box",
		Style: Style{
			Fg: "#ff0000",
			Bg: "#00ff00",
			Hover: &StateStyle{
				Fg: "#0000ff",
			},
		},
	})

	if box.Style.fgColor != (Color{R: 0xff, Set: true}) {
		t.Fatalf("fgColor = %+v", box.Style.fgColor)
	}
	if box.Style.bgColor != (Color{G: 0xff, Set: true}) {
		t.Fatalf("bgColor = %+v", box.Style.bgColor)
	}
	if box.Style.Hover.fgColor != (Color{B: 0xff, Set: true}) {
		t.Fatalf("hover fgColor = %+v", box.Style.Hover.fgColor)
	}
}

func TestInvalidColorPanics(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on invalid color")
		}
	}()
	NewBox(BoxProps{ID: "box", Style: Style{Fg: "not-a-color"}})
}
