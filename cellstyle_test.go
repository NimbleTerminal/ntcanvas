package ntcanvas

import "testing"

func TestCompactStyleTransitions(t *testing.T) {
	full := CellStyle{Foreground: RGBColor(255, 254, 253), Background: RGBColor(252, 251, 250), Attributes: Bold | Dim | Italic | Underline | Blink | Reverse | Strikethrough}
	var w runWriter
	w.cell("a", full)
	w.cell("b", CellStyle{Foreground: IndexedColor(255)})
	w.cell("c", CellStyle{})
	w.reset()
	want := "\x1b[1;2;3;4;5;7;9;38;2;255;254;253;48;2;252;251;250ma\x1b[0;38;5;255mb\x1b[0mc"
	if w.String() != want {
		t.Fatalf("%q != %q", w.String(), want)
	}
	w.cell("d", CellStyle{Attributes: 1 << 15, Foreground: Color(0xff000000)})
	if w.String() != want+"d" {
		t.Fatal("unsupported style emitted SGR")
	}
	w.cell("e", full)
	if w.style != full {
		t.Fatal("style transition")
	}
	// With output capacity available, even maximum-size style changes allocate nothing.
	var warm runWriter
	warm.Grow(2048)
	if n := testing.AllocsPerRun(10, func() {
		// Copying a nonzero Builder is invalid; reuse it while writing a bounded run.
		warm.transition(full)
	}); n != 0 {
		t.Fatalf("per-transition allocations: %g", n)
	}
}
