package ntcanvas

import (
	"image"

	"charm.land/lipgloss/v2"
	"nimbleterminal.dev/ntcanvas/internal/cellstore"
)

// FillRect fills r intersected with Bounds. It does not alter the default style.
func (m *StyledRuneGrid) FillRect(r image.Rectangle, c Cell) {
	if c.Rune != 0 {
		c.Rune = cellstore.PlainRune(c.Rune)
	}
	r = r.Intersect(m.cells.Area)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			m.cells.Content[y][x] = c
		}
	}
}

// ClearRect resets r intersected with Bounds to empty, unstyled cells.
func (m *StyledRuneGrid) ClearRect(r image.Rectangle) { m.FillRect(r, Cell{}) }

// SetStyleRect replaces styles in r intersected with Bounds, preserving runes
// and the default Style used for future writes. It does not merge attributes.
func (m *StyledRuneGrid) SetStyleRect(r image.Rectangle, s lipgloss.Style) {
	r = r.Intersect(m.cells.Area)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			m.cells.Content[y][x].Style = s
		}
	}
}

// CopyRect copies src's cells in r to dst, mapping r.Min to dst. Both canvases
// clip the copy without changing that mapping. Blank cells overwrite destination
// cells; no transparency or style merging is applied. Grids have no viewport state.
// Overlapping storage (including copies within the same grid) behaves as if the source
// cells were snapshotted first. Cell/style values are copied, not callback state.
// The temporary allocation is proportional to the clipped copied area.
func (m *StyledRuneGrid) CopyRect(dst Point, src *StyledRuneGrid, r image.Rectangle) {
	m.cells.CopyRect(dst, &src.cells, r)
}
