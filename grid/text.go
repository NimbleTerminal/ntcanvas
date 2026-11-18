package ntcanvasgrid

import (
	"strings"

	"nimbleterminal.dev/ntcanvas/internal/cellstore"
)

// SanitizeText replaces control, format, and line/paragraph separator runes with
// U+FFFD, including ESC, C0/C1, DEL, tabs, newlines, bidi controls, and joiners.
// Invalid UTF-8 also becomes U+FFFD. This neutralizes terminal controls, but does
// not interpret/remove ANSI sequences or make Unicode glyphs one column wide.
func SanitizeText(s string) string { return strings.Map(cellstore.PlainRune, s) }
