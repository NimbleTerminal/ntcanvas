package cellstore

import (
	"unicode"
	"unicode/utf8"

	"github.com/charmbracelet/x/ansi"
)

// PlainRune replaces a rune that could steer a terminal with U+FFFD.
func PlainRune(r rune) rune {
	// Also neutralize bidi/format controls and Unicode line/paragraph separators.
	if unicode.IsControl(r) || unicode.In(r, unicode.Cf, unicode.Zl, unicode.Zp) {
		return utf8.RuneError
	}
	return r
}

// SingleColumnRune is a rendering policy, not a storage conversion. Empty cells
// stay blank; all other runes must be sanitized and one column under Charm's rules.
func SingleColumnRune(r rune) rune {
	if r == 0 {
		return 0
	}
	r = PlainRune(r)
	if r >= ' ' && r <= '~' {
		return r
	}
	if ansi.StringWidth(string(r)) != 1 || AttachesAfterBoundary(string(r)) {
		return utf8.RuneError
	}
	return r
}

// NeedsGraphemeBoundary detects cells that would merge when concatenated.
// SGR style changes alone do not establish a grapheme boundary.
func NeedsGraphemeBoundary(previous, text string) bool {
	if previous == "" || len(previous) == 1 && len(text) == 1 {
		return false
	}
	cluster, _ := ansi.FirstGraphemeCluster(previous+text, ansi.GraphemeWidth)
	return len(cluster) > len(previous)
}

// AttachesAfterBoundary identifies leading marks that absorb a cell even with
// our ZWNJ separator. Their standalone width is not safe when concatenated.
func AttachesAfterBoundary(text string) bool {
	if len(text) == 1 {
		return false
	}
	return NeedsGraphemeBoundary(" \u200c", text)
}

// CopyAxis clips one axis of a rectangle copy. Unsigned distances avoid signed
// overflow for rectangles spanning MinInt to MaxInt. Conversion back to int
// happens only after clipping to a canvas axis.
func CopyAxis(dst, lo, hi, srcSize, dstSize int) (src, target, length int) {
	src, end := max(0, lo), min(srcSize, hi)
	if src >= end || dstSize == 0 || dst >= dstSize {
		return 0, 0, 0
	}
	offset := uint(src) - uint(lo)
	if offset > uint(dstSize-1)-uint(dst) {
		return 0, 0, 0
	}
	if dst < 0 {
		before := uint(0) - uint(dst)
		if offset < before {
			skip := before - offset
			if skip >= uint(end-src) {
				return 0, 0, 0
			}
			src += int(skip)
			offset += skip
		}
	}
	target = int(uint(dst) + offset)
	return src, target, min(end-src, dstSize-target)
}
