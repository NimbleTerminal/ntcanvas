package ntcanvas

import (
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"strings"
	"unicode"
	"unicode/utf8"
)

func plainRune(r rune) rune {
	// Also neutralize bidi/format controls and Unicode line/paragraph separators.
	if unicode.IsControl(r) || unicode.In(r, unicode.Cf, unicode.Zl, unicode.Zp) {
		return utf8.RuneError
	}
	return r
}

// SanitizeText replaces control, format, and line/paragraph separator runes with
// U+FFFD, including ESC, C0/C1, DEL, tabs, newlines, bidi controls, and joiners.
// Invalid UTF-8 also becomes U+FFFD. This neutralizes terminal controls, but does
// not interpret/remove ANSI sequences or make Unicode glyphs one column wide.
func SanitizeText(s string) string { return strings.Map(plainRune, s) }

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
				r = plainRune(r)
			}
			m.content[p.Y][x] = NewCellWithStyle(r, style)
		}
		x++
	}
	return true
}

// singleColumnRune is a rendering policy, not a storage conversion. Empty cells
// stay blank; all other runes must be sanitized and one column under Charm's rules.
func singleColumnRune(r rune) rune {
	if r == 0 {
		return 0
	}
	r = plainRune(r)
	if r >= ' ' && r <= '~' {
		return r
	}
	if ansi.StringWidth(string(r)) != 1 {
		return utf8.RuneError
	}
	return r
}

// needsGraphemeBoundary detects cells that would merge when concatenated.
// SGR style changes alone do not establish a grapheme boundary.
func needsGraphemeBoundary(previous, text string) bool {
	if previous == "" || len(previous) == 1 && len(text) == 1 {
		return false
	}
	cluster, _ := ansi.FirstGraphemeCluster(previous+text, ansi.GraphemeWidth)
	return len(cluster) > len(previous)
}
