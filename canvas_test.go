// ntcharts - Copyright (c) 2024 Neomantra Corp.

package ntcanvas

import (
	"testing"
)

func TestNew(t *testing.T) {
	w := 30
	h := 15
	c := NewStyledRuneGrid(w, h)
	m := NewModel(c)

	p := m.Cursor()
	if (p.X != 0) || (p.Y != 0) {
		t.Error("Cursor not initialized to (0,0)")
	}

	if c.Width() != w {
		t.Errorf("Width not initialized:%d", w)
	}
	if c.Height() != h {
		t.Errorf("Height not initialized:%d", h)
	}

	if m.ViewWidth != w {
		t.Errorf("ViewWidth not initialized:%d", w)
	}
	if m.ViewHeight != h {
		t.Errorf("ViewHeight not initialized:%d", h)
	}

}

func TestCursor(t *testing.T) {
	w := 30
	h := 15
	c := NewStyledRuneGrid(w, h)
	m := NewModel(c)

	m.SetCursor(Point{X: w / 2, Y: h / 2})
	p := m.Cursor()
	if p.X != w/2 {
		t.Errorf("CursorX not set correctly:%d", p.X)
	}
	if p.Y != h/2 {
		t.Errorf("CursorY not set correctly:%d", p.Y)
	}

	c.Clear()
	p = m.Cursor()
	if p.X != w/2 {
		t.Errorf("CursorX not set correctly after clear:%d", p.X)
	}
	if p.Y != h/2 {
		t.Errorf("CursorY not set correctly after clear:%d", p.Y)
	}

	m.SetCursor(Point{X: -1, Y: -1})
	p = m.Cursor()
	if p.X != 0 {
		t.Errorf("CursorX not bounded:%d", p.X)
	}
	if p.Y != 0 {
		t.Errorf("CursorY not bounded:%d", p.Y)
	}

	m.SetCursor(Point{X: w, Y: h})
	p = m.Cursor()
	if p.X != w-1 {
		t.Errorf("CursorX not bounded:%d", p.X)
	}
	if p.Y != h-1 {
		t.Errorf("CursorY not bounded:%d", p.Y)
	}
}

// TestResize_NegativeClampedToZero verifies Resize never lets negative dims
// reach the underlying make([]CellLine, h) call, which would panic. Higher
// layers (heatmap, picture) can pass negative values when their windowsize
// math underflows; clamping here removes a class of crashes from any
// upstream user of canvas.
func TestResize_NegativeClampedToZero(t *testing.T) {
	c := NewStyledRuneGrid(20, 10)
	c.Resize(-5, -3)
	if w, h := c.Width(), c.Height(); w != 0 || h != 0 {
		t.Errorf("expected Resize(-5,-3) to clamp to 0x0, got %dx%d", w, h)
	}
}

func TestResize(t *testing.T) {
	w := 30
	h := 15
	c := NewStyledRuneGrid(w, h)
	m := NewModel(c)
	m.SetCursor(Point{X: w / 2, Y: h / 2})

	nW := w + 5
	nH := h + 5
	c.Resize(nW, nH)

	p := m.Cursor()
	if (p.X != w/2) || (p.Y != h/2) {
		t.Error("Grid resize unexpectedly changed model cursor")
	}

	m.SetCursor(Point{X: w, Y: h})
	p = m.Cursor()
	if p.X != w {
		t.Errorf("CursorX not set correctly after resize:%d", p.X)
	}
	if p.Y != h {
		t.Errorf("CursorY not set correctly after resize:%d", p.Y)
	}
}

func TestSetRune(t *testing.T) {
	w := 10
	h := 5
	c := NewStyledRuneGrid(w, h)

	r := c.Cell(Point{X: w / 2, Y: h / 2}).Rune
	if r != 0 {
		t.Errorf("Rune not initialized:'%c'", r)
	}

	if c.SetRune(Point{X: w, Y: h}, 'A') {
		t.Error("SetRune not bounded")
	}

	if !c.SetRune(Point{X: w / 2, Y: h / 2}, 'A') {
		t.Error("SetRune not set correctly")
	}

	r = c.Cell(Point{X: w / 2, Y: h / 2}).Rune
	if r != 'A' {
		t.Errorf("Rune not set correctly:'%c'", r)
	}

	c.Clear()
	r = c.Cell(Point{X: w / 2, Y: h / 2}).Rune
	if r != 0 {
		t.Errorf("Rune not clear correctly:'%c'", r)
	}
}

func TestFill(t *testing.T) {
	w := 10
	h := 5
	c := NewStyledRuneGrid(w, h)

	c.Fill(NewCell('B'))

	for y := 0; y < c.Height(); y++ {
		for x := 0; x < c.Width(); x++ {
			r := c.Cell(Point{X: x, Y: y}).Rune
			if r != 'B' {
				t.Errorf("Rune did not fill correctly:'%c'", r)
				return
			}
		}
	}
}

func TestFloat64Point(t *testing.T) {
	x := -1.5
	y := 2.5
	f := Float64Point{x, y}

	scale := 2.0
	sc := Float64Point{scale, scale}
	nf := f.Mul(sc)
	if nf.X != x*scale {
		t.Errorf("Float64Point X value did not Mul correctly:%f", nf.X)
	}
	if nf.Y != y*scale {
		t.Errorf("Float64Point Y value did not Mul correctly:%f", nf.Y)
	}

	xOffset := -10.0
	yOffset := 3.0
	offset := Float64Point{X: xOffset, Y: yOffset}
	nf = f.Add(offset)
	if nf.X != x+xOffset {
		t.Errorf("Float64Point X value did not Add correctly:%f", nf.X)
	}
	if nf.Y != y+yOffset {
		t.Errorf("Float64Point Y value did not Add correctly:%f", nf.Y)
	}

	offset = Float64Point{X: xOffset, Y: yOffset}
	nf = f.Sub(offset)
	if nf.X != x-xOffset {
		t.Errorf("Float64Point X value did not Sub correctly:%f", nf.X)
	}
	if nf.Y != y-yOffset {
		t.Errorf("Float64Point Y value did not Sub correctly:%f", nf.Y)
	}
}
