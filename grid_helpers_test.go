package ntcanvas

import (
	"fmt"
	"image"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestUnchangedResizeDoesNotAllocate(t *testing.T) {
	type grid interface {
		Surface
		Resize(int, int)
		ResizeChecked(int, int, SizeLimits) error
		SetText(Point, string) bool
	}
	for _, size := range []Point{{}, {X: 3}, {Y: 3}, {X: 80, Y: 24}} {
		for _, g := range []grid{
			NewRuneGrid(size.X, size.Y),
			NewStyledRuneGrid(size.X, size.Y),
			NewGraphemeGrid(size.X, size.Y),
		} {
			t.Run(fmt.Sprintf("%T/%dx%d", g, size.X, size.Y), func(t *testing.T) {
				g.SetText(Point{}, "A界e\u0301")
				before := g.Render(g.Bounds())
				var err error
				allocs := testing.AllocsPerRun(10, func() {
					g.Resize(size.X, size.Y)
					err = g.ResizeChecked(size.X, size.Y, SizeLimits{})
				})
				if err != nil || allocs != 0 {
					t.Fatalf("unchanged resize: err=%v allocations=%g", err, allocs)
				}
				if g.Bounds() != image.Rect(0, 0, size.X, size.Y) || g.Render(g.Bounds()) != before {
					t.Fatal("unchanged resize altered bounds or content")
				}
				// Checked calls must not bypass validation just because sizes match.
				if err := g.ResizeChecked(size.X, size.Y, SizeLimits{MaxWidth: -1}); err != ErrInvalidSize {
					t.Fatal("invalid limits accepted", err)
				}
				if size.X > 1 {
					if err := g.ResizeChecked(size.X, size.Y, SizeLimits{MaxWidth: 1}); err != ErrSizeLimit {
						t.Fatal("stricter limit accepted", err)
					}
				}
				if g.Render(g.Bounds()) != before {
					t.Fatal("rejected resize altered content")
				}
			})
		}
	}
}

func TestRenderAllocationBudgets(t *testing.T) {
	styled := NewStyledRuneGrid(40, 12)
	grapheme := NewGraphemeGrid(80, 24)
	grapheme.Fill("x", CellStyle{Foreground: IndexedColor(1)})
	runeGrid := NewRuneGrid(80, 24)
	runeGrid.Fill(RuneCell{Rune: 'x', Style: CellStyle{Foreground: IndexedColor(1)}})
	for _, tc := range []struct {
		name string
		grid Surface
		max  float64
	}{
		{"styled blanks", styled, 4}, {"ASCII graphemes", grapheme, 16}, {"compact runes", runeGrid, 16},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var result string
			n := testing.AllocsPerRun(10, func() { result = tc.grid.Render(tc.grid.Bounds()) })
			if result == "" || n > tc.max {
				t.Fatalf("allocations %g > %g (output length %d)", n, tc.max, len(result))
			}
		})
	}
}

func TestGridConveniences(t *testing.T) {
	type grid interface {
		Surface
		Width() int
		Height() int
		SetText(Point, string) bool
		SetLines([]string)
		ShiftLeft()
		ShiftRight()
		ShiftUp()
		ShiftDown()
	}
	for _, g := range []grid{NewStyledRuneGrid(3, 2), NewRuneGrid(3, 2), NewGraphemeGrid(3, 2)} {
		if g.Width() != 3 || g.Height() != 2 {
			t.Fatal("dimensions")
		}
		g.SetLines([]string{"abc", "def", "ignored"})
		g.ShiftLeft()
		g.ShiftRight()
		g.ShiftUp()
		g.ShiftDown()
		if got := g.Render(g.Bounds()); got != "   \n ef" {
			t.Fatalf("%T: %q", g, got)
		}
		g.SetText(Point{}, "\x00\x1b\n")
		if got := g.Render(image.Rect(0, 0, 3, 1)); got != "���" {
			t.Fatalf("%T controls: %q", g, got)
		}
	}
	for _, g := range []grid{NewStyledRuneGrid(0, 2), NewRuneGrid(0, 2), NewGraphemeGrid(0, 2), NewRuneGrid(2, 0), NewGraphemeGrid(2, 0)} {
		g.ShiftLeft()
		g.ShiftRight()
		g.ShiftUp()
		g.ShiftDown()
		if g.Render(g.Bounds()) != "" {
			t.Fatalf("%T: empty grid", g)
		}
	}
}

func TestStyledSafeWrites(t *testing.T) {
	s := lipgloss.NewStyle().Bold(true)
	for name, write := range map[string]func(*StyledRuneGrid){
		"text":        func(g *StyledRuneGrid) { g.SetText(Point{}, "\x1b") },
		"styled text": func(g *StyledRuneGrid) { g.SetTextWithStyle(Point{}, "\x1b", s) },
		"lines":       func(g *StyledRuneGrid) { g.SetLinesWithStyle([]string{"\x1b"}, s) },
		"runes":       func(g *StyledRuneGrid) { g.SetRunesWithStyle(Point{}, []rune{'\x1b'}, s) },
		"rune":        func(g *StyledRuneGrid) { g.SetRuneWithStyle(Point{}, '\x1b', s) },
		"cell":        func(g *StyledRuneGrid) { g.SetCell(Point{}, NewCell('\x1b')) },
		"fill":        func(g *StyledRuneGrid) { g.Fill(NewCell('\x1b')) },
		"line":        func(g *StyledRuneGrid) { g.FillLine(0, NewCell('\x1b')) },
		"rect":        func(g *StyledRuneGrid) { g.FillRect(g.Bounds(), NewCell('\x1b')) },
	} {
		t.Run(name, func(t *testing.T) {
			g := NewStyledRuneGrid(2, 1)
			write(g)
			if g.Cell(Point{}).Rune != '�' {
				t.Fatal("unsanitized")
			}
		})
	}
	g := NewStyledRuneGrid(1, 1, WithContent([]string{"\x1b"}))
	if g.Cell(Point{}).Rune != '�' {
		t.Fatal("content option")
	}
	g.SetRawTextWithStyle(Point{}, "\x1b", s)
	if g.Cell(Point{}).Rune != '\x1b' || !g.Cell(Point{}).Style.GetBold() {
		t.Fatal("raw text changed")
	}
	g.SetCell(Point{}, Cell{})
	if g.Cell(Point{}).Rune != 0 {
		t.Fatal("zero cell should clear")
	}
}
