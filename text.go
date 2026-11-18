package ntcanvas

import (
	"charm.land/lipgloss/v2"
	"nimbleterminal.dev/ntcanvas/internal/cellstore"
)

// SetText writes text using the default style. Controls are
// replaced as in SanitizeText, one stored cell per rune. The
// return value reports a valid row, not whether any cells were written.
// Sanitization does not preserve display columns; prefer GraphemeGrid for untrusted text.
func (m *StyledRuneGrid) SetText(p Point, s string) bool {
	return m.SetTextWithStyle(p, s, m.Style)
}

// SetTextWithStyle is SetText with an explicit, trusted style. This
// sanitizes text only: a style transform or border may itself emit arbitrary
// output, and padding/borders/width rules may expand a cell's rendered footprint.
func (m *StyledRuneGrid) SetTextWithStyle(p Point, s string, style lipgloss.Style) bool {
	return m.writeString(p, s, style, true)
}

func (m *StyledRuneGrid) writeString(p Point, s string, style lipgloss.Style, plain bool) bool {
	if !m.insideYBounds(p.Y) {
		return false
	}
	width := m.Width()
	if p.X >= width || (p.X < 0 && p.X <= -len(s)) {
		return true
	}
	x := p.X
	for _, r := range s {
		if x >= width {
			break
		}
		if x >= 0 {
			if plain {
				r = cellstore.PlainRune(r)
			}
			m.cells.Content[p.Y][x] = NewCellWithStyle(r, style)
		}
		x++
	}
	return true
}
