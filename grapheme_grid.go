package ntcanvas

import (
	"image"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
)

// MaxGraphemeRunes bounds stored and rendered cluster complexity. Longer clusters
// are replaced wholesale with one U+FFFD cell, not split into multiple clusters.
const MaxGraphemeRunes = 16

func boundedGrapheme(text string, width int) (string, int) {
	count := 0
	for range text {
		count++
		if count > MaxGraphemeRunes {
			return string(utf8.RuneError), 1
		}
	}
	return text, width
}

// GraphemeCell is a read-only snapshot returned by Cell. Width is 1 or 2 at a
// leading cell, 0 at a continuation. Empty cells have empty Text and Width 1.
// A continuation belongs to the preceding width-2 cell and shares its style.
type GraphemeCell struct {
	Text         string
	Style        CellStyle
	Width        int
	Continuation bool
}

// GraphemeGrid stores clusters in terminal columns, using Charm ANSI's grapheme/width rules.
// Terminals/fonts may differ. Joined emoji and combining sequences stay intact.
// Mutations never leave orphan continuations. Do not copy grid values or mutate
// from multiple goroutines. Stored cluster strings are owned by this grid.
type GraphemeGrid struct {
	storage[GraphemeCell]
	Style CellStyle
}

func NewGraphemeGrid(w, h int) *GraphemeGrid { g := &GraphemeGrid{}; g.Resize(w, h); return g }
func NewGraphemeGridChecked(w, h int, l SizeLimits) (*GraphemeGrid, error) {
	if err := ValidateSize(w, h, l); err != nil {
		return nil, err
	}
	return NewGraphemeGrid(w, h), nil
}
func (g *GraphemeGrid) Bounds() image.Rectangle { return g.area }
func (g *GraphemeGrid) Resize(w, h int) {
	g.resize(w, h)
	// A width-2 lead truncated at the right boundary becomes an empty cell.
	if w = g.area.Dx(); w > 0 {
		for _, row := range g.content {
			if row[w-1].Width == 2 {
				row[w-1] = GraphemeCell{}
			}
		}
	}
}
func (g *GraphemeGrid) ResizeChecked(w, h int, l SizeLimits) error {
	if err := ValidateSize(w, h, l); err != nil {
		return err
	}
	g.Resize(w, h)
	return nil
}
func (g *GraphemeGrid) Clear() { g.clear() }
func (g *GraphemeGrid) Cell(p Point) GraphemeCell {
	if !p.In(g.area) {
		return GraphemeCell{}
	}
	c := g.content[p.Y][p.X]
	if c.Text == "" && !c.Continuation {
		c.Width = 1
	}
	return c
}

func graphemeText(s string) string {
	// Keep joiners for grapheme shaping, while neutralizing other format/control
	// runes. Unlike rune-grid sanitation, this preserves joined emoji sequences.
	return strings.Map(func(r rune) rune {
		if r == '\u200d' || r == '\u200c' {
			return r
		}
		return plainRune(r)
	}, s)
}
func (g *GraphemeGrid) erase(p Point) {
	if !p.In(g.area) {
		return
	}
	row := g.content[p.Y]
	x := p.X
	if row[x].Continuation {
		x--
	}
	width := max(1, row[x].Width)
	for i := x; i < min(x+width, len(row)); i++ {
		row[i] = GraphemeCell{}
	}
}
func (g *GraphemeGrid) put(p Point, text string, width int, style CellStyle) bool {
	if !p.In(g.area) || width < 1 || width > 2 || width > g.area.Dx()-p.X {
		return false
	}
	for x := p.X; x < p.X+width; x++ {
		g.erase(Point{X: x, Y: p.Y})
	}
	g.content[p.Y][p.X] = GraphemeCell{Text: strings.Clone(text), Style: style, Width: width}
	if width == 2 {
		g.content[p.Y][p.X+1] = GraphemeCell{Style: style, Continuation: true}
	}
	return true
}

// SetGrapheme writes exactly one sanitized, positive-width cluster. Invalid,
// multiple, zero-width or partially clipped clusters leave the grid unchanged.
// A cluster over MaxGraphemeRunes becomes a width-1 replacement character.
func (g *GraphemeGrid) SetGrapheme(p Point, text string, style CellStyle) bool {
	text = graphemeText(text)
	cluster, width := ansi.FirstGraphemeCluster(text, ansi.GraphemeWidth)
	if len(cluster) != len(text) {
		return false
	}
	cluster, width = boundedGrapheme(cluster, width)
	return g.put(p, cluster, width, style)
}

// SetText advances by display columns. Partially clipped clusters are skipped;
// standalone zero-width clusters are discarded (never attached to old content).
// The result indicates a valid row. Controls become U+FFFD; joiners are retained.
// Clusters over MaxGraphemeRunes are replaced wholesale with U+FFFD.
func (g *GraphemeGrid) SetText(p Point, text string) bool {
	return g.SetTextWithStyle(p, text, g.Style)
}
func (g *GraphemeGrid) SetTextWithStyle(p Point, text string, style CellStyle) bool {
	if p.Y < 0 || p.Y >= g.area.Dy() {
		return false
	}
	if p.X >= g.area.Dx() || (p.X < 0 && p.X <= -len(text)) {
		return true
	}
	text = graphemeText(text)
	for len(text) > 0 {
		if p.X >= g.area.Dx() {
			break
		}
		cluster, width := ansi.FirstGraphemeCluster(text, ansi.GraphemeWidth)
		if len(cluster) == 0 {
			break
		}
		text = text[len(cluster):]
		cluster, width = boundedGrapheme(cluster, width)
		if width < 1 || width > 2 {
			continue
		}
		g.put(p, cluster, width, style)
		if p.X >= 0 && width >= g.area.Dx()-p.X {
			break
		}
		p.X += width
	}
	return true
}

// ClearRect clears every cluster touched by r, including its other half just
// outside r. This is necessary to preserve valid wide-character storage.
func (g *GraphemeGrid) ClearRect(r image.Rectangle) {
	r = r.Intersect(g.area)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			g.erase(Point{X: x, Y: y})
		}
	}
}

// SetStyleRect replaces styles on entire clusters touched by r, including a
// cluster's other half outside r. Empty cells in r are styled as well.
func (g *GraphemeGrid) SetStyleRect(r image.Rectangle, s CellStyle) {
	r = r.Intersect(g.area)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			lead := x
			if g.content[y][x].Continuation {
				lead--
			}
			n := max(1, g.content[y][lead].Width)
			for i := lead; i < lead+n; i++ {
				g.content[y][i].Style = s
			}
		}
	}
}

// CopyRect snapshots the clipped region. A source cluster cut by either clipping
// boundary becomes blank cells. Overwriting any destination half clears its
// entire old cluster, which may clear a cell just outside the destination area.
func (g *GraphemeGrid) CopyRect(dst Point, src *GraphemeGrid, r image.Rectangle) {
	sx, dx, w := copyAxis(dst.X, r.Min.X, r.Max.X, src.area.Dx(), g.area.Dx())
	sy, dy, h := copyAxis(dst.Y, r.Min.Y, r.Max.Y, src.area.Dy(), g.area.Dy())
	if w == 0 || h == 0 {
		return
	}
	cells := make([]GraphemeCell, w*h)
	for y := 0; y < h; y++ {
		copy(cells[y*w:(y+1)*w], src.content[sy+y][sx:sx+w])
	}
	g.ClearRect(image.Rect(dx, dy, dx+w, dy+h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := cells[y*w+x]
			p := Point{X: dx + x, Y: dy + y}
			if c.Continuation || c.Width == 2 && x+1 >= w {
				continue
			}
			if c.Text == "" {
				g.content[p.Y][p.X] = GraphemeCell{Style: c.Style}
				continue
			}
			g.put(p, c.Text, c.Width, c.Style)
		}
	}
}

// Render emits blanks for any cluster bisected by the requested rectangle.
func (g *GraphemeGrid) Render(r image.Rectangle) string {
	r = r.Intersect(g.area)
	if r.Empty() {
		return ""
	}
	var w runWriter
	for y := r.Min.Y; y < r.Max.Y; y++ {
		if y > r.Min.Y {
			w.WriteByte('\n')
		}
		previous := ""
		for x := r.Min.X; x < r.Max.X; x++ {
			c := g.content[y][x]
			text, style, width := c.Text, c.Style, max(1, c.Width)
			if c.Continuation || c.Width == 2 && x+1 >= r.Max.X {
				text, style, width = " ", CellStyle{}, 1
			}
			if text == "" {
				text = " "
			}
			// Independently written regional indicators or trailing joiners can merge
			// when concatenated. ZWNJ preserves the explicit cell boundary without a
			// display column or terminal cursor command. SGR alone does not do this.
			// Sanitized single-byte ASCII cells cannot merge.
			if needsGraphemeBoundary(previous, text) {
				w.WriteRune('\u200c')
			}
			w.cell(text, style)
			previous = text
			x += width - 1
		}
		w.reset()
	}
	return w.String()
}

// FillRect tiles one sanitized grapheme across r, clearing touched old clusters.
// An odd final column for a width-2 tile stays blank. Returns false for invalid
// clusters without changing the grid. As with ClearRect, boundary halves of old
// clusters outside r are cleared to preserve integrity.
func (g *GraphemeGrid) FillRect(r image.Rectangle, text string, style CellStyle) bool {
	text = graphemeText(text)
	cluster, width := ansi.FirstGraphemeCluster(text, ansi.GraphemeWidth)
	if len(cluster) != len(text) {
		return false
	}
	cluster, width = boundedGrapheme(cluster, width)
	if width < 1 || width > 2 {
		return false
	}
	r = r.Intersect(g.area)
	g.ClearRect(r)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; width <= r.Max.X-x; x += width {
			g.put(Point{X: x, Y: y}, cluster, width, style)
		}
	}
	return true
}
