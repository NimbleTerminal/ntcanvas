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
	Rune  rune
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
func (m *StyledRuneGrid) ResizeChecked(w, h int, limits SizeLimits) error {
	if err := ValidateSize(w, h, limits); err != nil {
		return err
	}
	m.Resize(w, h)
	return nil
}
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

func (m *StyledRuneGrid) Width() int {
	return m.area.Dx()
}

func (m *StyledRuneGrid) Height() int {
	return m.area.Dy()
}

func (m *StyledRuneGrid) Clear() {
	for i := range m.content {
		clear(m.content[i])
	}
}

func (m *StyledRuneGrid) SetLines(lines []string) {
	m.SetLinesWithStyle(lines, m.Style)
}

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

func (m *StyledRuneGrid) SetRunes(p Point, l []rune) bool {
	return m.SetRunesWithStyle(p, l, m.Style)
}

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

func (m *StyledRuneGrid) SetRune(p Point, r rune) bool {
	return m.SetCell(p, NewCellWithStyle(r, m.Style))
}
func (m *StyledRuneGrid) SetRuneWithStyle(p Point, r rune, style lipgloss.Style) bool {
	return m.SetCell(p, NewCellWithStyle(r, style))
}

func (m *StyledRuneGrid) Cell(p Point) (c Cell) {
	if !p.In(m.area) {
		return
	}
	c = m.content[p.Y][p.X]
	return
}

func (m *StyledRuneGrid) Fill(c Cell) { m.FillRect(m.Bounds(), c) }
func (m *StyledRuneGrid) FillLine(y int, c Cell) {
	if !m.insideYBounds(y) {
		return
	}
	m.FillRect(image.Rect(0, y, m.Width(), y+1), c)
}

func (m *StyledRuneGrid) SetStyle(s lipgloss.Style) {
	for i := range m.content {
		for j := range m.content[i] {
			m.content[i][j].Style = s
		}
	}
}

// Shifts move stored cells one position and clear the exposed edge.
func (m *StyledRuneGrid) ShiftUp()    { m.shiftUp() }
func (m *StyledRuneGrid) ShiftDown()  { m.shiftDown() }
func (m *StyledRuneGrid) ShiftLeft()  { m.shiftLeft() }
func (m *StyledRuneGrid) ShiftRight() { m.shiftRight() }

func (m *StyledRuneGrid) insideYBounds(y int) bool {
	if (y < 0) || (y >= m.area.Dy()) {
		return false
	}
	return true
}
