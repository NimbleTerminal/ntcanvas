// screenframes emits actual canvas output for the headless terminal test.
// Expected screens live separately in testdata/screen; this tool never writes them.
package main

import (
	"encoding/json"
	"image"
	"math"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"nimbleterminal.dev/ntcanvas"
)

type frame struct {
	Name   string `json:"name"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Output string `json:"output"`
}

func main() {
	c := ntcanvas.NewStyledRuneGrid(8, 3, ntcanvas.WithContent([]string{"ABCDEFGH", "ijklmnop", "12345678"}))
	m := ntcanvas.NewModel(c)
	m.SetViewport(image.Rect(0, 0, 4, 2))
	c.SetStyleRect(image.Rect(0, 0, 2, 1), lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000")).Bold(true))
	var frames []frame
	capture := func(name string, w, h int) {
		// A minimal full-frame host: position at home, erase row tails and the
		// remainder of the screen, and use CRLF for terminal line movement.
		// Canvas itself returns content; it does not own cursor placement/erasure.
		output := "\x1b[H" + strings.ReplaceAll(m.View(), "\n", "\x1b[K\r\n") + "\x1b[K\x1b[J"
		frames = append(frames, frame{name, w, h, output})
	}
	capture("initial", 8, 4)
	overlay := ntcanvas.NewStyledRuneGrid(3, 1, ntcanvas.WithContent([]string{"x y"}))
	overlay.SetStyle(lipgloss.NewStyle().Background(lipgloss.Color("#0000ff")))
	c.CopyRect(ntcanvas.Point{X: 1, Y: 1}, overlay, overlay.Bounds())
	capture("composed", 8, 4)
	m.SetCursor(ntcanvas.Point{X: 1, Y: 1})
	capture("panned", 8, 4)
	c.ClearRect(image.Rect(1, 1, 3, 2))
	capture("cleared", 8, 4)
	c.Resize(3, 2)
	m.SetViewport(c.Bounds())
	capture("resized", 5, 3)
	c.Clear()
	c.SetText(ntcanvas.Point{}, "\x1bX")
	capture("safe-text", 5, 3)
	// The same viewport wrapper renders both new compact surfaces.
	runes := ntcanvas.NewRuneGrid(5, 2)
	runes.SetTextWithStyle(ntcanvas.Point{}, "abc", ntcanvas.CellStyle{Foreground: ntcanvas.RGBColor(255, 0, 0), Attributes: ntcanvas.Bold})
	m = ntcanvas.NewModel(runes)
	capture("compact-runes", 8, 4)
	g := ntcanvas.NewGraphemeGrid(12, 2)
	g.SetText(ntcanvas.Point{}, "Ae\u0301界👩‍💻🇺🇸Z")
	m = ntcanvas.NewModel(g)
	capture("graphemes", 12, 3)
	g.SetText(ntcanvas.Point{X: 3}, "X")
	capture("wide-overwrite", 12, 3)
	m.SetViewport(image.Rect(5, 0, 9, 1))
	capture("wide-clipping", 12, 3)
	g.Clear()
	g.SetText(ntcanvas.Point{}, "🇺")
	g.SetText(ntcanvas.Point{X: 2}, "🇸")
	m.SetViewport(image.Rect(0, 0, 8, 1))
	capture("cluster-boundary", 12, 3)

	runes = ntcanvas.NewRuneGrid(10, 1)
	runes.SingleColumn = true
	runes.SetText(ntcanvas.Point{}, "界e\u0301👩‍💻")
	m = ntcanvas.NewModel(runes)
	capture("single-column-runes", 12, 2)
	c = ntcanvas.NewStyledRuneGrid(10, 1)
	c.SingleColumn = true
	c.SetRawText(ntcanvas.Point{}, "界e\u0301👩‍💻")
	m = ntcanvas.NewModel(c)
	capture("single-column-styled", 12, 2)
	g = ntcanvas.NewGraphemeGrid(10, 1)
	g.SetText(ntcanvas.Point{}, "a"+strings.Repeat("\u0301", 100000)+"👩"+strings.Repeat("\u200d👩", 60)+"Z")
	m = ntcanvas.NewModel(g)
	capture("bounded-clusters", 12, 2)

	g = ntcanvas.NewGraphemeGrid(8, 3)
	ntcanvas.RasterLine(g.Bounds(), ntcanvas.Float64Point{X: -1e30, Y: 1}, ntcanvas.Float64Point{X: 1e30, Y: 1}, func(p ntcanvas.Point) { g.SetGrapheme(p, "·", ntcanvas.CellStyle{}) })
	ntcanvas.RasterLine(g.Bounds(), ntcanvas.Float64Point{}, ntcanvas.Float64Point{X: 2, Y: 2}, func(p ntcanvas.Point) { g.SetGrapheme(p, "•", ntcanvas.CellStyle{}) })
	m = ntcanvas.NewModel(g)
	capture("clipped-lines", 10, 4)

	runes = ntcanvas.NewRuneGrid(3, 2)
	ntcanvas.RasterLineAA(runes.Bounds(), ntcanvas.Float64Point{X: 0, Y: 0.5}, ntcanvas.Float64Point{X: 2, Y: 0.5}, func(p ntcanvas.Point, coverage float64) {
		level := uint8(math.Round(255 * coverage))
		runes.SetRuneWithStyle(p, '·', ntcanvas.CellStyle{Foreground: ntcanvas.RGBColor(level, level, level), Background: ntcanvas.RGBColor(0, 0, 0)})
	})
	m = ntcanvas.NewModel(runes)
	capture("antialiased-line", 5, 3)
	if err := json.NewEncoder(os.Stdout).Encode(frames); err != nil {
		panic(err)
	}
}
