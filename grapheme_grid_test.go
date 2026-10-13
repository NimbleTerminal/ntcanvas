package ntcanvas

import (
	"image"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
)

func assertClusters(t *testing.T, g *GraphemeGrid) {
	t.Helper()
	for y := 0; y < g.area.Dy(); y++ {
		for x := 0; x < g.area.Dx(); x++ {
			c := g.Cell(Point{X: x, Y: y})
			if c.Continuation {
				if x == 0 {
					t.Fatal("orphan at left edge")
				}
				p := g.Cell(Point{X: x - 1, Y: y})
				if p.Width != 2 || p.Continuation || p.Style != c.Style {
					t.Fatal("orphan", x, y)
				}
			} else if c.Text != "" {
				if utf8.RuneCountInString(c.Text) > MaxGraphemeRunes {
					t.Fatal("oversized cluster")
				}
				if ansi.StringWidth(c.Text) != c.Width {
					t.Fatal("width", c)
				}
				if c.Width == 2 && (x+1 >= g.area.Dx() || !g.Cell(Point{X: x + 1, Y: y}).Continuation) {
					t.Fatal("missing continuation", x, y)
				}
			}
		}
	}
	if !g.Bounds().Empty() {
		for _, row := range strings.Split(g.Render(g.Bounds()), "\n") {
			if ansi.StringWidth(row) != g.area.Dx() {
				t.Fatalf("render width: %q", row)
			}
		}
	}
}
func TestGraphemeText(t *testing.T) {
	g := NewGraphemeGrid(12, 1)
	g.SetText(Point{}, "Ae\u0301界👩‍💻🇺🇸Z")
	for _, tc := range []struct {
		x     int
		text  string
		width int
	}{{0, "A", 1}, {1, "e\u0301", 1}, {2, "界", 2}, {4, "👩‍💻", 2}, {6, "🇺🇸", 2}, {8, "Z", 1}} {
		c := g.Cell(Point{X: tc.x})
		if c.Text != tc.text || c.Width != tc.width {
			t.Fatal(tc, c)
		}
	}
	assertClusters(t, g)
	if g.SetGrapheme(Point{}, "ab", CellStyle{}) || g.SetGrapheme(Point{}, "\u0301", CellStyle{}) || g.SetGrapheme(Point{X: 11}, "界", CellStyle{}) {
		t.Fatal("invalid cluster accepted")
	}
	g.Clear()
	g.SetText(Point{X: -1}, "界ab")
	if g.Render(g.Bounds()) != " ab         " {
		t.Fatalf("left clipping %q", g.Render(g.Bounds()))
	}
	g.Clear()
	g.SetText(Point{}, "\u0301a\x1b\u202e\n")
	if !strings.HasPrefix(g.Render(g.Bounds()), "a���") {
		t.Fatal(g.Render(g.Bounds()))
	}
	g.Clear()
	g.SetText(Point{}, "界Z")
	if g.Render(image.Rect(1, 0, 3, 1)) != " Z" || g.Render(image.Rect(0, 0, 1, 1)) != " " {
		t.Fatal("partial render")
	}
}
func TestGraphemeOverwriteResizeAndStyle(t *testing.T) {
	g := NewGraphemeGrid(5, 1)
	g.SetText(Point{}, "界語x")
	g.SetText(Point{X: 1}, "A") // overwrites a continuation, clearing the old lead
	if g.Render(g.Bounds()) != " A語x" {
		t.Fatal(g.Render(g.Bounds()))
	}
	g.SetText(Point{X: 1}, "界") // clears both cells of the overlapped next cluster
	if g.Render(g.Bounds()) != " 界 x" {
		t.Fatal(g.Render(g.Bounds()))
	}
	s := CellStyle{Foreground: IndexedColor(1)}
	g.SetStyleRect(image.Rect(2, 0, 3, 1), s)
	if g.Cell(Point{X: 1}).Style != s || g.Cell(Point{X: 2}).Style != s {
		t.Fatal("half style")
	}
	g.Resize(2, 1)
	if g.Render(g.Bounds()) != "  " {
		t.Fatal("truncated cluster", g.Render(g.Bounds()))
	}
	g.Resize(5, 1)
	g.SetText(Point{}, "界語x")
	g.ClearRect(image.Rect(1, 0, 3, 1))
	if g.Render(g.Bounds()) != "    x" {
		t.Fatal("clear partial clusters")
	}
	g.FillRect(g.Bounds(), "界", CellStyle{})
	if g.Render(g.Bounds()) != "界界 " {
		t.Fatal("tile", g.Render(g.Bounds()))
	}
	if g.FillRect(g.Bounds(), "ab", CellStyle{}) {
		t.Fatal("invalid fill")
	}
	assertClusters(t, g)
}
func TestGraphemeCopy(t *testing.T) {
	for _, tc := range []struct {
		dst  int
		src  image.Rectangle
		want string
	}{
		{1, image.Rect(0, 0, 4, 1), " 界語 "},
		{0, image.Rect(1, 0, 4, 1), " 語 x "},
		{0, image.Rect(0, 0, 3, 1), "界  x "},
		{-1, image.Rect(0, 0, 4, 1), " 語 x "},
	} {
		g := NewGraphemeGrid(6, 1)
		g.SetText(Point{}, "界語x")
		g.CopyRect(Point{X: tc.dst}, g, tc.src)
		if got := g.Render(g.Bounds()); got != tc.want {
			t.Fatalf("%+v got %q", tc, got)
		}
		assertClusters(t, g)
	}
}
func FuzzGraphemeMutations(f *testing.F) {
	for _, s := range []string{"界語x", "e\u0301👩‍💻", "\x1b[2J", "🇺🇸", "a" + strings.Repeat("\u0301", 17)} {
		f.Add(s, uint8(1), uint8(4))
	}
	f.Fuzz(func(t *testing.T, s string, a, b uint8) {
		if len(s) > 1024 {
			t.Skip()
		}
		g := NewGraphemeGrid(8, 2)
		g.SetText(Point{}, s)
		assertClusters(t, g)
		x, y := int(a%10)-1, int(b%10)-1
		g.CopyRect(Point{X: x, Y: 1}, g, image.Rect(y, 0, 7, 1))
		assertClusters(t, g)
		g.SetStyleRect(image.Rect(x, 0, 7, 2), CellStyle{Attributes: Bold})
		assertClusters(t, g)
		g.SetText(Point{X: y}, "界")
		assertClusters(t, g)
		g.ClearRect(image.Rect(x, 0, 6, 1))
		assertClusters(t, g)
		for _, shift := range []func(){g.ShiftLeft, g.ShiftRight, g.ShiftUp, g.ShiftDown} {
			shift()
			assertClusters(t, g)
		}
		g.Resize(int(a%9), int(b%3))
		assertClusters(t, g)
	})
}

func TestIndependentClustersDoNotMergeOnRender(t *testing.T) {
	for _, pair := range [][2]string{{"🇺", "🇸"}, {"👩‍", "💻"}, {"🇺", "🇸🇨"}} {
		g := NewGraphemeGrid(8, 1)
		g.SetText(Point{}, pair[0])
		g.SetText(Point{X: 2}, pair[1])
		if !strings.Contains(g.Render(g.Bounds()), "\u200c") {
			t.Fatal("missing cluster boundary")
		}
		assertClusters(t, g)
	}
}
