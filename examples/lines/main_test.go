package main

import (
	tea "charm.land/bubbletea/v2"
	"nimbleterminal.dev/ntcanvas"
	"testing"
)

func TestLineControlsAndCoverage(t *testing.T) {
	m := newModel()
	n, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m = n.(model)
	n, _ = m.Update(tea.KeyPressMsg{Code: 'g', Text: "g"})
	m = n.(model)
	if m.step != 1 || !m.blocks {
		t.Fatal(m.Summary())
	}
	g := lineGrid(24, 8, 0, false, true)
	partial := false
	full := ntcanvas.RGBColor(141, 227, 182)
	for y := 0; y < g.Height(); y++ {
		for x := 0; x < g.Width(); x++ {
			c := g.Cell(ntcanvas.Point{X: x, Y: y})
			if c.Rune == '·' && c.Style.Foreground != full {
				partial = true
			}
		}
	}
	if !partial {
		t.Fatal("AA did not produce blended foregrounds")
	}
	for _, size := range []int{0, 1, 12, 40, 80, int(^uint(0) >> 1)} {
		n, _ = m.Update(tea.WindowSizeMsg{Width: size, Height: size})
		m = n.(model)
		m.View()
	}
}
