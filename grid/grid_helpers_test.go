package ntcanvasgrid

import "testing"

func TestCompactFillAndStyles(t *testing.T) {
	s := CellStyle{Attributes: Bold}
	r := NewRuneGrid(3, 2)
	r.Fill(RuneCell{Rune: 'a'})
	r.FillLine(1, RuneCell{Rune: 'b', Style: s})
	r.FillLine(-1, RuneCell{Rune: '!'})
	r.FillLine(int(^uint(0)>>1), RuneCell{Rune: '!'})
	if r.Cell(Point{Y: 1}) != (RuneCell{Rune: 'b', Style: s}) || r.Cell(Point{}).Rune != 'a' {
		t.Fatal("rune fill")
	}
	r.SetStyle(s)
	r.ShiftRight()
	r.ShiftDown()
	if r.Cell(Point{X: 1, Y: 1}) != (RuneCell{Rune: 'a', Style: s}) {
		t.Fatal("rune shift style")
	}
	r.SetLinesWithStyle([]string{"x"}, CellStyle{})
	if r.Cell(Point{}).Style != (CellStyle{}) {
		t.Fatal("line style")
	}
	g := NewGraphemeGrid(3, 2)
	if !g.Fill("界", s) || !g.FillLine(1, "x", CellStyle{}) || g.FillLine(-1, "!", s) || g.FillLine(int(^uint(0)>>1), "!", s) {
		t.Fatal("grapheme fill result")
	}
	if g.Cell(Point{}).Text != "界" || g.Cell(Point{X: 2}).Text != "" || g.Cell(Point{Y: 1}).Text != "x" {
		t.Fatal("grapheme fill cells")
	}
	g.SetStyle(s)
	g.ShiftRight()
	g.ShiftDown()
	if g.Cell(Point{X: 1, Y: 1}).Style != s {
		t.Fatal("grapheme shift style")
	}
	assertClusters(t, g)
	g.SetLinesWithStyle([]string{"x"}, CellStyle{})
	if g.Cell(Point{}).Style != (CellStyle{}) {
		t.Fatal("grapheme line style")
	}
}

func TestGraphemeShiftEdges(t *testing.T) {
	for _, direction := range []string{"left", "right"} {
		g := NewGraphemeGrid(4, 1)
		g.SetText(Point{}, "界語")
		if direction == "left" {
			g.ShiftLeft()
		} else {
			g.ShiftRight()
		}
		want := " 語 "
		if direction == "right" {
			want = " 界 "
		}
		if got := g.Render(g.Bounds()); got != want {
			t.Fatalf("%s: %q", direction, got)
		}
		assertClusters(t, g)
	}
}
