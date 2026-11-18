package ntcanvas

import (
	"image"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestSingleColumnRenderingPolicy(t *testing.T) {
	for _, input := range []string{"界界界界界", "a\u0301\u0301\u0301", "👩‍💻", "🇺🇸", "\x1b\n\t", "abc"} {
		r := NewRuneGrid(10, 1)
		s := NewStyledRuneGrid(10, 1)
		r.SetText(Point{}, input)
		s.SetRawText(Point{}, input)
		beforeR, beforeS := r.Cell(Point{}), s.Cell(Point{})
		r.SingleColumn, s.SingleColumn = true, true
		for _, g := range []Surface{r, s} {
			out := g.Render(g.Bounds())
			if width := ansi.StringWidth(out); width != 10 {
				t.Fatalf("%T %q: width %d", g, input, width)
			}
			if strings.ContainsAny(out, "\x1b\n\t") {
				t.Fatalf("controls survived: %q", out)
			}
		}
		if r.Cell(Point{}) != beforeR || s.Cell(Point{}).Rune != beforeS.Rune {
			t.Fatal("render changed stored text")
		}
	}
	src := NewRuneGrid(2, 1)
	src.SetText(Point{}, "界\u0301")
	dst := NewRuneGrid(2, 1)
	dst.SingleColumn = true
	dst.CopyRect(Point{}, src, src.Bounds())
	if dst.Render(dst.Bounds()) != "��" {
		t.Fatal("copied content bypassed rendering policy")
	}
	s := NewStyledRuneGrid(2, 1)
	s.SetText(Point{}, "界x")
	if ansi.StringWidth(s.Render(s.Bounds())) != 3 {
		t.Fatal("default behavior changed")
	}
}

func TestSingleColumnCellBoundaries(t *testing.T) {
	for _, input := range []string{"\u1161\u1161\u11a8", "\u0d4ea", "\U000111c2a", "a\u0d4e"} {
		r := NewRuneGrid(5, 2)
		s := NewStyledRuneGrid(5, 2)
		r.SingleColumn, s.SingleColumn = true, true
		for y := range 2 {
			r.SetText(Point{Y: y}, input)
			s.SetText(Point{Y: y}, input)
		}
		for _, g := range []Surface{r, s} {
			for _, rect := range []image.Rectangle{g.Bounds(), image.Rect(1, 0, 4, 2)} {
				out := g.Render(rect)
				for _, row := range strings.Split(out, "\n") {
					if got := ansi.StringWidth(row); got != rect.Dx() {
						t.Fatalf("%T input %q rect %v: %q width %d, want %d", g, input, rect, row, got, rect.Dx())
					}
					if strings.HasPrefix(row, "\u200c") {
						t.Fatalf("%T carried a boundary across rows or clipping: %q", g, row)
					}
				}
			}
		}
		if r.Cell(Point{}).Rune != []rune(input)[0] || s.Cell(Point{}).Rune != []rune(input)[0] {
			t.Fatal("render changed stored runes")
		}
		r.SingleColumn, s.SingleColumn = false, false
		for _, g := range []Surface{r, s} {
			if strings.Contains(g.Render(g.Bounds()), "\u200c") {
				t.Fatalf("%T added boundaries with SingleColumn disabled", g)
			}
		}
	}
}

// These positive-width standalone marks still attach after a ZWNJ separator.
func TestStandaloneSpacingMarks(t *testing.T) {
	for _, input := range []string{"\u17c7", "\u1a55", "\u1929", "\U000114bb", "\u1a57", "\u1ad1", "\u0ac0", "\u1a71", "\u1a6e"} {
		t.Run(input, func(t *testing.T) {
			g := NewGraphemeGrid(8, 2)
			r := NewRuneGrid(8, 2)
			s := NewStyledRuneGrid(8, 2)
			r.SingleColumn, s.SingleColumn = true, true
			for y := range 2 {
				g.SetText(Point{X: 1, Y: y}, input)
				r.SetText(Point{X: 1, Y: y}, input)
				s.SetText(Point{X: 1, Y: y}, input)
			}
			for _, grid := range []Surface{g, r, s} {
				for _, rect := range []image.Rectangle{grid.Bounds(), image.Rect(1, 0, 3, 2)} {
					for _, row := range strings.Split(grid.Render(rect), "\n") {
						if width := ansi.StringWidth(row); width != rect.Dx() {
							t.Fatalf("%T %q: width %d, want %d", grid, row, width, rect.Dx())
						}
					}
				}
			}
			if g.Cell(Point{X: 1}).Text != input || r.Cell(Point{X: 1}).Rune != []rune(input)[0] || s.Cell(Point{X: 1}).Rune != []rune(input)[0] {
				t.Fatal("render changed stored text")
			}
		})
	}
}
