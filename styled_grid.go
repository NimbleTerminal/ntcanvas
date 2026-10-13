// ntcharts - Copyright (c) 2024 Neomantra Corp.

package ntcanvas

import (
	"charm.land/lipgloss/v2"
	"image"
	"strings"
)

var defaultStyle = lipgloss.NewStyle()

// Cell contains a rune and lipgloss Style for rendering
type Cell struct {
	// Rune is the stored code point; zero renders as a space.
	Rune rune
	// Style is a trusted Lip Gloss style and may expand the rendered cell.
	Style lipgloss.Style
}

// NewCell returns Cell with given rune and default lipgloss Style.
func NewCell(r rune) Cell {
	return Cell{Rune: r, Style: defaultStyle}
}

// NewCellWithStyle returns Cell with given rune and lipgloss Style.
func NewCellWithStyle(r rune, s lipgloss.Style) Cell {
	return Cell{Rune: r, Style: s}
}

// CellLine is a slice of Cells
type CellLine []Cell

// StyledRuneGrid is a mutable rune grid with full Lip Gloss cell styles.
// Writes advance one cell per rune. Wide/combining runes and layout styles can
// expand rendered output. Use one owning event loop; do not copy grid values.
type StyledRuneGrid struct {
	storage[Cell]
	// Style is the default style for future writes, not existing cells.
	Style lipgloss.Style
	// SingleColumn replaces non-one-column runes during rendering. Styles remain trusted.
	SingleColumn bool
}

// StyledOption configures a StyledRuneGrid. Options execute in order.
type StyledOption func(*StyledRuneGrid)

// NewStyledRuneGrid allocates an uncapped grid. Negative sizes clamp to zero.
func NewStyledRuneGrid(w, h int, opts ...StyledOption) *StyledRuneGrid {
	g := &StyledRuneGrid{Style: defaultStyle}
	g.Resize(w, h)
	for _, opt := range opts {
		opt(g)
	}
	return g
}

// NewStyledRuneGridChecked validates dimensions before allocating.
func NewStyledRuneGridChecked(w, h int, limits SizeLimits, opts ...StyledOption) (*StyledRuneGrid, error) {
	if err := ValidateSize(w, h, limits); err != nil {
		return nil, err
	}
	return NewStyledRuneGrid(w, h, opts...), nil
}

// Resize preserves overlapping cells; models viewing this grid keep their state.
func (m *StyledRuneGrid) Resize(w, h int) { m.resize(w, h) }

// ResizeChecked validates dimensions before resizing. A rejected request
// leaves the grid unchanged.
func (m *StyledRuneGrid) ResizeChecked(w, h int, limits SizeLimits) error {
	if err := ValidateSize(w, h, limits); err != nil {
		return err
	}
	m.Resize(w, h)
	return nil
}

// Bounds returns the grid rectangle, with its origin at (0, 0).
func (m *StyledRuneGrid) Bounds() image.Rectangle { return m.area }

// Render clips to Bounds and joins visible rows without a trailing newline.
func (m *StyledRuneGrid) Render(r image.Rectangle) string {
	r = r.Intersect(m.area)
	if r.Empty() {
		return ""
	}
	var b strings.Builder
	b.Grow(r.Dx() * r.Dy())
	for y := r.Min.Y; y < r.Max.Y; y++ {
		if y > r.Min.Y {
			b.WriteByte('\n')
		}
		previous := ""
		for x := r.Min.X; x < r.Max.X; x++ {
			c := m.content[y][x]
			if m.SingleColumn {
				c.Rune = singleColumnRune(c.Rune)
			}
			text := " "
			if c.Rune != 0 {
				text = string(c.Rune)
			}
			if m.SingleColumn && needsGraphemeBoundary(previous, text) {
				b.WriteRune('\u200c')
			}
			b.WriteString(c.Style.Render(text))
			previous = text
		}
	}
	return b.String()
}

// Width returns the number of stored columns.
func (m *StyledRuneGrid) Width() int {
	return m.area.Dx()
}

// Height returns the number of stored rows.
func (m *StyledRuneGrid) Height() int {
	return m.area.Dy()
}

// Clear resets every cell to empty content and a zero style, preserving dimensions
// and the default Style used for future writes.
func (m *StyledRuneGrid) Clear() {
	for i := range m.content {
		clear(m.content[i])
	}
}

// SetLines writes each string to a row starting at (0, 0), using the default
// Style and SetText semantics. Extra rows are clipped; unwritten cells remain.
func (m *StyledRuneGrid) SetLines(lines []string) {
	m.SetLinesWithStyle(lines, m.Style)
}

// SetLinesWithStyle is SetLines with an explicit style. It preserves cells
// beyond each written line and clips rows outside Bounds.
func (m *StyledRuneGrid) SetLinesWithStyle(lines []string, s lipgloss.Style) {
	for y, l := range lines {
		if y >= m.area.Dy() {
			break
		}
		m.SetTextWithStyle(Point{X: 0, Y: y}, l, s)
	}
}

// SetRawText preserves terminal controls. Use SetText for external text.
func (m *StyledRuneGrid) SetRawText(p Point, l string) bool {
	return m.SetRawTextWithStyle(p, l, m.Style)
}

// SetRawTextWithStyle preserves controls and uses a trusted Lip Gloss style.
func (m *StyledRuneGrid) SetRawTextWithStyle(p Point, l string, s lipgloss.Style) bool {
	return m.writeString(p, l, s, false)
}

// SetRunes writes one cell per rune using the default Style, sanitizing controls.
// It returns false for an invalid row and true even when all runes are clipped.
func (m *StyledRuneGrid) SetRunes(p Point, l []rune) bool {
	return m.SetRunesWithStyle(p, l, m.Style)
}

// SetRunesWithStyle is SetRunes with an explicit, trusted Lip Gloss style.
func (m *StyledRuneGrid) SetRunesWithStyle(p Point, l []rune, s lipgloss.Style) bool {
	if !m.insideYBounds(p.Y) {
		return false
	}
	if p.X >= m.Width() || (p.X < 0 && p.X <= -len(l)) {
		return true
	}
	start := 0
	if p.X < 0 {
		start = -p.X
	}
	x := max(0, p.X)
	for _, r := range l[start : start+min(len(l)-start, m.Width()-x)] {
		m.content[p.Y][x] = NewCellWithStyle(plainRune(r), s)
		x++
	}
	return true
}

// SetCell writes c at p, replacing control runes with U+FFFD. A zero rune
// clears the content while retaining c.Style. It returns false outside Bounds.
func (m *StyledRuneGrid) SetCell(p Point, c Cell) bool {
	if !p.In(m.area) {
		return false
	}
	if c.Rune != 0 {
		c.Rune = plainRune(c.Rune)
	}
	m.content[p.Y][p.X] = c
	return true
}

// SetCellStyle replaces the style at p without changing its rune.
// It returns false outside Bounds. Styles and their callbacks remain trusted.
func (m *StyledRuneGrid) SetCellStyle(p Point, s lipgloss.Style) bool {
	if !p.In(m.area) {
		return false
	}
	m.content[p.Y][p.X].Style = s
	return true
}

// GetCellStyle returns a pointer to a copy, or nil outside the grid.
// Application-provided styles and their callbacks remain trusted.
func (m *StyledRuneGrid) GetCellStyle(p Point) *lipgloss.Style {
	if !p.In(m.area) {
		return nil
	}
	copyStyle := m.content[p.Y][p.X].Style
	return &copyStyle
}

// SetRune writes r at p with the default Style, replacing control runes with
// U+FFFD. A zero rune clears the content. It returns false outside Bounds.
func (m *StyledRuneGrid) SetRune(p Point, r rune) bool {
	return m.SetCell(p, NewCellWithStyle(r, m.Style))
}

// SetRuneWithStyle writes r at p with an explicit style, replacing control
// runes with U+FFFD. A zero rune clears content. It returns false outside Bounds.
func (m *StyledRuneGrid) SetRuneWithStyle(p Point, r rune, style lipgloss.Style) bool {
	return m.SetCell(p, NewCellWithStyle(r, style))
}

// Cell returns a copy of the cell at p, or a zero cell outside Bounds.
func (m *StyledRuneGrid) Cell(p Point) (c Cell) {
	if !p.In(m.area) {
		return
	}
	c = m.content[p.Y][p.X]
	return
}

// Fill sets every cell to c, sanitizing nonzero control runes.
// It leaves the default Style unchanged.
func (m *StyledRuneGrid) Fill(c Cell) { m.FillRect(m.Bounds(), c) }

// FillLine fills row y with c, sanitizing nonzero control runes.
// Rows outside Bounds are ignored; the default Style is unchanged.
func (m *StyledRuneGrid) FillLine(y int, c Cell) {
	if !m.insideYBounds(y) {
		return
	}
	m.FillRect(image.Rect(0, y, m.Width(), y+1), c)
}

// SetStyle replaces every cell style without changing content or the default
// Style used for future writes.
func (m *StyledRuneGrid) SetStyle(s lipgloss.Style) {
	for i := range m.content {
		for j := range m.content[i] {
			m.content[i][j].Style = s
		}
	}
}

// ShiftUp moves stored cells up one row and clears the bottom row.
func (m *StyledRuneGrid) ShiftUp() { m.shiftUp() }

// ShiftDown moves stored cells down one row and clears the top row.
func (m *StyledRuneGrid) ShiftDown() { m.shiftDown() }

// ShiftLeft moves stored cells left one column and clears the right edge.
func (m *StyledRuneGrid) ShiftLeft() { m.shiftLeft() }

// ShiftRight moves stored cells right one column and clears the left edge.
func (m *StyledRuneGrid) ShiftRight() { m.shiftRight() }

func (m *StyledRuneGrid) insideYBounds(y int) bool {
	if (y < 0) || (y >= m.area.Dy()) {
		return false
	}
	return true
}
