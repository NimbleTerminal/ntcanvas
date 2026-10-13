package ntcanvas

import (
	"strconv"
	"strings"
)

// Color is a terminal color. Its zero value selects the terminal default.
// Construct nondefault colors with IndexedColor or RGBColor.
type Color uint32

func IndexedColor(index uint8) Color { return Color(0x1000000 | uint32(index)) }
func RGBColor(r, g, b uint8) Color {
	return Color(0x2000000 | uint32(r)<<16 | uint32(g)<<8 | uint32(b))
}

// Attributes are terminal SGR flags, not layout or text transformations.
type Attributes uint16

const (
	Bold Attributes = 1 << iota
	Dim
	Italic
	Underline
	Blink
	Reverse
	Strikethrough
)

// CellStyle is a compact, comparable style shared by RuneGrid and GraphemeGrid.
// The zero value uses default colors and no attributes. Unsupported bits/colors
// are ignored during rendering. There are no borders, padding, or callbacks.
type CellStyle struct {
	Foreground, Background Color
	Attributes             Attributes
}

// runWriter coalesces equal terminal styles; each row ends in the default style.
// This optimization is deliberately not used for arbitrary Lip Gloss styles.
type runWriter struct {
	strings.Builder
	style CellStyle
}

func (w *runWriter) cell(text string, s CellStyle) {
	if s != w.style {
		w.transition(s)
		w.style = s
	}
	w.WriteString(text)
}
func (w *runWriter) reset() {
	if w.style != (CellStyle{}) {
		w.WriteString("\x1b[0m")
		w.style = CellStyle{}
	}
}

// transition builds SGR parameters on the stack. When replacing a style, reset
// and new attributes share a single sequence so removed attributes cannot leak.
func (w *runWriter) transition(s CellStyle) {
	var buf [64]byte // reset, seven attributes, and two RGB colors fit within 64 bytes
	codes := buf[:0]
	add := func(n int) {
		if len(codes) > 0 {
			codes = append(codes, ';')
		}
		codes = strconv.AppendInt(codes, int64(n), 10)
	}
	if w.style != (CellStyle{}) {
		add(0)
	}
	for i, code := range [...]int{1, 2, 3, 4, 5, 7, 9} {
		if s.Attributes&(1<<i) != 0 {
			add(code)
		}
	}
	for i, c := range [...]Color{s.Foreground, s.Background} {
		prefix := 38 + 10*i
		switch uint32(c) >> 24 {
		case 1:
			add(prefix)
			add(5)
			add(int(c & 255))
		case 2:
			add(prefix)
			add(2)
			add(int(c >> 16 & 255))
			add(int(c >> 8 & 255))
			add(int(c & 255))
		}
	}
	if len(codes) != 0 {
		w.WriteString("\x1b[")
		w.Write(codes)
		w.WriteByte('m')
	}
}
