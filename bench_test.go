package ntcanvas_test

import (
	"charm.land/lipgloss/v2"
	"nimbleterminal.dev/ntcanvas"
	"strings"
	"testing"
)

var sink string

func BenchmarkViewSmall(b *testing.B) {
	c := ntcanvas.NewStyledRuneGrid(40, 12)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		sink = c.Render(c.Bounds())
	}
}
func BenchmarkViewLargeBacking(b *testing.B) {
	c := ntcanvas.NewModel(ntcanvas.NewStyledRuneGrid(400, 300), ntcanvas.WithViewWidth(40), ntcanvas.WithViewHeight(12))
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		sink = c.View()
	}
}
func BenchmarkViewStyled(b *testing.B) {
	c := ntcanvas.NewStyledRuneGrid(80, 24)
	c.Fill(ntcanvas.NewCellWithStyle('x', lipgloss.NewStyle().Foreground(lipgloss.Color("1"))))
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		sink = c.Render(c.Bounds())
	}
}
func BenchmarkClear(b *testing.B) {
	c := ntcanvas.NewStyledRuneGrid(80, 24)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		c.Clear()
	}
}
func BenchmarkClippedText(b *testing.B) {
	c := ntcanvas.NewStyledRuneGrid(40, 1)
	s := strings.Repeat("x", 1000000)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		c.SetText(ntcanvas.Point{}, s)
	}
}

func BenchmarkRuneGridStyled(b *testing.B) {
	g := ntcanvas.NewRuneGrid(80, 24)
	g.FillRect(g.Bounds(), ntcanvas.RuneCell{Rune: 'x', Style: ntcanvas.CellStyle{Foreground: ntcanvas.IndexedColor(1)}})
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		sink = g.Render(g.Bounds())
	}
}
func BenchmarkGraphemeGridStyled(b *testing.B) {
	g := ntcanvas.NewGraphemeGrid(80, 24)
	g.FillRect(g.Bounds(), "x", ntcanvas.CellStyle{Foreground: ntcanvas.IndexedColor(1)})
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		sink = g.Render(g.Bounds())
	}
}
