package main

import (
	tea "charm.land/bubbletea/v2"
	"nimbleterminal.dev/ntcanvas"
	"testing"
)

func TestGridControls(t *testing.T) {
	m := newModel()
	send := func(k string) { n, _ := m.Update(tea.KeyPressMsg{Code: rune(k[0]), Text: k}); m = n.(model) }
	send("s")
	if !m.styled.SingleColumn || !m.runes.SingleColumn {
		t.Fatal(m.Summary())
	}
	send("c")
	if m.graphemes.Cell(ntcanvas.Point{X: 6, Y: 1}).Text != "" || m.graphemes.Cell(ntcanvas.Point{X: 8, Y: 1}).Text != "語" {
		t.Fatal("wide glyph clearing")
	}
	if m.styled.Cell(ntcanvas.Point{X: 6, Y: 1}).Rune != '界' || m.runes.Cell(ntcanvas.Point{X: 7, Y: 1}).Rune != 0 {
		t.Fatal("rune clearing")
	}
	send("r")
	if m.safe || m.cleared || m.graphemes.Cell(ntcanvas.Point{X: 6, Y: 1}).Text != "界" {
		t.Fatal(m.Summary())
	}
	for _, size := range []int{0, 1, 12, 40, 80, int(^uint(0) >> 1)} {
		n, _ := m.Update(tea.WindowSizeMsg{Width: size, Height: size})
		m = n.(model)
		m.View()
	}
}
