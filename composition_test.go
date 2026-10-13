package ntcanvas

import (
	"image"
	"testing"

	"charm.land/lipgloss/v2"
)

func TestCopyRectClippingAndOverlap(t *testing.T) {
	// Compare small copies with a simple per-cell snapshot oracle. Include
	// shallow aliases, all overlap directions, empty/inverted rectangles, and
	// clipping on each edge. The oracle's bounded coordinates cannot overflow.
	for _, shared := range []bool{false, true} {
		for dx := -4; dx <= 4; dx++ {
			for dy := -3; dy <= 3; dy++ {
				for x := -2; x <= 4; x++ {
					for y := -1; y <= 3; y++ {
						src := NewStyledRuneGrid(4, 3, WithContent([]string{"abcd", "efgh", "ijkl"}))
						src.SetCell(Point{X: 1, Y: 1}, Cell{})
						src.SetCellStyle(Point{X: 2, Y: 1}, lipgloss.NewStyle().Bold(true))
						dst := NewStyledRuneGrid(4, 3, WithContent([]string{"....", "....", "...."}))
						if shared {
							dst = src
						}
						var want [3][4]Cell
						for yy := range want {
							for xx := range want[yy] {
								want[yy][xx] = dst.Cell(Point{X: xx, Y: yy})
							}
						}
						r := image.Rectangle{Min: Point{X: x, Y: y}, Max: Point{X: 3, Y: 2}}
						for yy := r.Min.Y; yy < r.Max.Y; yy++ {
							for xx := r.Min.X; xx < r.Max.X; xx++ {
								p, q := Point{X: xx, Y: yy}, Point{X: dx + xx - x, Y: dy + yy - y}
								if p.In(src.Bounds()) && q.In(dst.Bounds()) {
									want[q.Y][q.X] = src.Cell(p)
								}
							}
						}
						dst.CopyRect(Point{X: dx, Y: dy}, src, r)
						for yy := range want {
							for xx, c := range want[yy] {
								got := dst.Cell(Point{X: xx, Y: yy})
								if got.Rune != c.Rune || got.Style.GetBold() != c.Style.GetBold() {
									t.Fatalf("shared=%v dst=(%d,%d) rect=%v at (%d,%d): got %q want %q", shared, dx, dy, r, xx, yy, got.Rune, c.Rune)
								}
							}
						}
					}
				}
			}
		}
	}
}

func TestCopyRectExtremeCoordinates(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	minInt := -maxInt - 1
	src := NewStyledRuneGrid(2, 1, WithContent([]string{"ab"}))
	for _, tc := range []struct {
		dst  Point
		r    image.Rectangle
		want string
	}{
		{Point{}, image.Rectangle{Min: Point{X: minInt}, Max: Point{X: maxInt, Y: 1}}, ".."},
		{Point{X: minInt}, image.Rectangle{Min: Point{X: minInt}, Max: Point{X: maxInt, Y: 1}}, "ab"},
		{Point{X: maxInt}, src.Bounds(), ".."},
		{Point{X: minInt}, src.Bounds(), ".."},
		{Point{Y: minInt}, image.Rectangle{Min: Point{Y: minInt}, Max: Point{X: 2, Y: maxInt}}, "ab"},
	} {
		dst := NewStyledRuneGrid(2, 1, WithContent([]string{".."}))
		dst.CopyRect(tc.dst, src, tc.r)
		if got := dst.Render(dst.Bounds()); got != tc.want {
			t.Fatalf("%+v: %q", tc, got)
		}
	}
	empty := NewStyledRuneGrid(0, 0)
	empty.CopyRect(Point{}, src, src.Bounds())
	src.CopyRect(Point{}, empty, image.Rect(0, 0, 2, 1))
	if src.Render(src.Bounds()) != "ab" {
		t.Fatal("empty source changed destination")
	}
}

func TestCompositionCopiesValues(t *testing.T) {
	src := NewStyledRuneGrid(2, 1, WithContent([]string{"ab"}))
	dst := NewStyledRuneGrid(3, 1)
	dst.Style = lipgloss.NewStyle().Italic(true)
	dst.CopyRect(Point{}, src, src.Bounds())
	src.SetRune(Point{}, 'z')
	if dst.Cell(Point{}).Rune != 'a' || !dst.Style.GetItalic() {
		t.Fatal("copy aliased cells or changed default style")
	}
}

func TestSetStyleRect(t *testing.T) {
	m := NewStyledRuneGrid(3, 2, WithContent([]string{"abc", "def"}))
	m.Style = lipgloss.NewStyle().Italic(true)
	m.SetStyle(lipgloss.NewStyle().Bold(true))
	m.SetStyleRect(image.Rect(-1, 0, 2, 4), lipgloss.NewStyle().Underline(true))
	for y := 0; y < 2; y++ {
		for x := 0; x < 3; x++ {
			c := m.Cell(Point{X: x, Y: y})
			if c.Rune != rune('a'+y*3+x) || c.Style.GetBold() != (x == 2) || c.Style.GetUnderline() != (x < 2) {
				t.Fatalf("cell (%d,%d)", x, y)
			}
		}
	}
	if !m.Style.GetItalic() {
		t.Fatal("default style changed")
	}
	m.SetStyleRect(image.Rectangle{Min: Point{X: 2, Y: 2}, Max: Point{X: 1, Y: 1}}, lipgloss.NewStyle())
	m.SetStyleRect(image.Rect(8, 8, 9, 9), lipgloss.NewStyle())
	if !m.Cell(Point{}).Style.GetUnderline() {
		t.Fatal("empty region changed style")
	}
}
