package ntcanvas

import "image"

// RuneCell is a rune and compact terminal style. Zero renders as a space.
type RuneCell struct {
	// Rune is the stored code point; zero renders as a space.
	Rune rune
	// Style controls terminal colors and attributes.
	Style CellStyle
}

// RuneGrid stores one rune per cell using terminal-only styles. Wide characters
// and combining marks retain one-cell storage semantics; use GraphemeGrid when
// layout must follow display columns. Controls are replaced on every write.
// Grids are pointer-owned and not safe for concurrent mutation; do not copy them.
type RuneGrid struct {
	storage[RuneCell]
	// Style is the default style for future writes, not existing cells.
	Style CellStyle
	// SingleColumn replaces non-one-column runes during rendering, without changing stored cells.
	SingleColumn bool
}

// NewRuneGrid allocates an uncapped rune grid. Negative dimensions clamp to
// zero; impossible allocation arithmetic panics with ErrInvalidSize. Use
// NewRuneGridChecked for externally supplied dimensions.
func NewRuneGrid(w, h int) *RuneGrid { g := &RuneGrid{}; g.Resize(w, h); return g }

// NewRuneGridChecked validates dimensions against limits before allocating.
func NewRuneGridChecked(w, h int, l SizeLimits) (*RuneGrid, error) {
	if err := ValidateSize(w, h, l); err != nil {
		return nil, err
	}
	return NewRuneGrid(w, h), nil
}

// Bounds returns the grid rectangle, with its origin at (0, 0).
func (g *RuneGrid) Bounds() image.Rectangle { return g.area }

// Resize preserves overlapping cells and clamps negative dimensions to zero.
// It panics with ErrInvalidSize if allocation arithmetic would overflow.
func (g *RuneGrid) Resize(w, h int) { g.resize(w, h) }

// ResizeChecked validates dimensions before resizing. A rejected request
// leaves the grid unchanged.
func (g *RuneGrid) ResizeChecked(w, h int, l SizeLimits) error {
	if err := ValidateSize(w, h, l); err != nil {
		return err
	}
	g.Resize(w, h)
	return nil
}

// Clear resets every cell to empty content and a zero style, preserving dimensions
// and the default Style used for future writes.
func (g *RuneGrid) Clear() { g.clear() }

// Cell returns a copy of the cell at p, or a zero cell outside Bounds.
func (g *RuneGrid) Cell(p Point) RuneCell {
	if !p.In(g.area) {
		return RuneCell{}
	}
	return g.content[p.Y][p.X]
}

// SetCell writes c at p, replacing control runes with U+FFFD. A zero rune
// clears the content while retaining c.Style. It returns false outside Bounds.
func (g *RuneGrid) SetCell(p Point, c RuneCell) bool {
	if !p.In(g.area) {
		return false
	}
	if c.Rune != 0 {
		c.Rune = plainRune(c.Rune)
	}
	g.content[p.Y][p.X] = c
	return true
}

// SetRune writes r at p with the default Style, replacing control runes with
// U+FFFD. A zero rune clears the content. It returns false outside Bounds.
func (g *RuneGrid) SetRune(p Point, r rune) bool { return g.SetCell(p, RuneCell{r, g.Style}) }

// SetRuneWithStyle writes a rune with an explicit style, sanitizing controls.
// A zero rune clears the cell's content while retaining the supplied style.
func (g *RuneGrid) SetRuneWithStyle(p Point, r rune, style CellStyle) bool {
	return g.SetCell(p, RuneCell{r, style})
}

// SetText writes one cell per rune. The result indicates a valid row, even if
// the entire text is clipped. It replaces terminal controls as in SanitizeText.
func (g *RuneGrid) SetText(p Point, s string) bool { return g.SetTextWithStyle(p, s, g.Style) }

// SetTextWithStyle is SetText with an explicit style. It returns false for an
// invalid row and true even when the text is fully clipped.
func (g *RuneGrid) SetTextWithStyle(p Point, s string, style CellStyle) bool {
	if p.Y < 0 || p.Y >= g.area.Dy() {
		return false
	}
	if p.X >= g.area.Dx() || (p.X < 0 && p.X <= -len(s)) {
		return true
	}
	for _, r := range s {
		if p.X >= g.area.Dx() {
			break
		}
		if p.X >= 0 {
			g.content[p.Y][p.X] = RuneCell{plainRune(r), style}
		}
		p.X++
	}
	return true
}

// FillRect fills r intersected with Bounds, sanitizing nonzero control runes.
// It leaves the default Style unchanged.
func (g *RuneGrid) FillRect(r image.Rectangle, c RuneCell) {
	if c.Rune != 0 {
		c.Rune = plainRune(c.Rune)
	}
	g.fillRect(r, c)
}

// ClearRect resets r intersected with Bounds to empty, unstyled cells.
func (g *RuneGrid) ClearRect(r image.Rectangle) { g.FillRect(r, RuneCell{}) }

// SetStyleRect replaces styles in r intersected with Bounds, preserving runes
// and the default Style used for future writes.
func (g *RuneGrid) SetStyleRect(r image.Rectangle, s CellStyle) {
	r = r.Intersect(g.area)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			g.content[y][x].Style = s
		}
	}
}

// CopyRect copies src cells in r to dst, mapping r.Min to dst and clipping
// to both grids without changing that mapping. Blank cells overwrite content.
// Overlaps use a snapshot; temporary storage is proportional to the copied area.
func (g *RuneGrid) CopyRect(dst Point, src *RuneGrid, r image.Rectangle) {
	g.copyRect(dst, &src.storage, r)
}

// Render clips r to Bounds and joins visible rows without a trailing newline.
// An empty intersection returns an empty string.
func (g *RuneGrid) Render(r image.Rectangle) string {
	r = r.Intersect(g.area)
	if r.Empty() {
		return ""
	}
	var w runWriter
	for y := r.Min.Y; y < r.Max.Y; y++ {
		if y > r.Min.Y {
			w.WriteByte('\n')
		}
		var previous rune
		for x := r.Min.X; x < r.Max.X; x++ {
			c := g.content[y][x]
			v := c.Rune
			if g.SingleColumn {
				v = singleColumnRune(v)
			}
			if v == 0 {
				v = ' '
			}
			if g.SingleColumn && previous != 0 && needsGraphemeBoundary(string(previous), string(v)) {
				w.WriteRune('\u200c')
			}
			w.cell(string(v), c.Style)
			previous = v
		}
		w.reset()
	}
	return w.String()
}
