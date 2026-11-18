package ntcanvasgrid

import (
	"image"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
)

func TestGraphemeComplexityLimit(t *testing.T) {
	within := "a" + strings.Repeat("\u0301", MaxGraphemeRunes-1)
	for _, input := range []string{within, within + "\u0301", "a" + strings.Repeat("\u0301", 100000), "👩" + strings.Repeat("\u200d👩", 60)} {
		want := input
		if utf8.RuneCountInString(input) > MaxGraphemeRunes {
			want = "�"
		}
		for _, write := range []func(*GraphemeGrid) bool{
			func(g *GraphemeGrid) bool { return g.SetGrapheme(Point{}, input, CellStyle{}) },
			func(g *GraphemeGrid) bool { return g.SetText(Point{}, input) },
			func(g *GraphemeGrid) bool { return g.FillRect(image.Rect(0, 0, 1, 1), input, CellStyle{}) },
		} {
			g := NewGraphemeGrid(3, 1)
			if !write(g) || g.Cell(Point{}).Text != want {
				t.Fatalf("cluster length %d not handled", utf8.RuneCountInString(input))
			}
			if got := g.Render(g.Bounds()); ansi.StringWidth(got) != 3 || len(got) > MaxGraphemeRunes*utf8.UTFMax+2 {
				t.Fatal("unbounded output")
			}
			assertClusters(t, g)
		}
	}
	g := NewGraphemeGrid(4, 1)
	long := "a" + strings.Repeat("\u0301", 100000)
	g.SetText(Point{}, long+"Z")
	if g.Render(g.Bounds()) != "�Z  " {
		t.Fatal("oversized cluster split or swallowed next cluster")
	}
	g.SetText(Point{X: -1}, long+"Y")
	if g.Cell(Point{}).Text != "Y" {
		t.Fatal("replacement clipping")
	}
	if g.SetGrapheme(Point{}, long+"Z", CellStyle{}) {
		t.Fatal("accepted multiple clusters")
	}
}
