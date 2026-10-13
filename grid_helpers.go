package ntcanvas

import "image"

// Width returns the number of stored columns.
func (g *RuneGrid) Width() int { return g.area.Dx() }

// Height returns the number of stored rows.
func (g *RuneGrid) Height() int { return g.area.Dy() }

// Width returns the number of stored columns.
func (g *GraphemeGrid) Width() int { return g.area.Dx() }

// Height returns the number of stored rows.
func (g *GraphemeGrid) Height() int { return g.area.Dy() }

// Fill sets every cell to c, sanitizing nonzero control runes.
// It leaves the default Style unchanged.
func (g *RuneGrid) Fill(c RuneCell) { g.FillRect(g.Bounds(), c) }

// FillLine fills row y with c, sanitizing nonzero control runes.
// Rows outside Bounds are ignored; the default Style is unchanged.
func (g *RuneGrid) FillLine(y int, c RuneCell) {
	if y < 0 || y >= g.Height() {
		return
	}
	g.FillRect(image.Rect(0, y, g.Width(), y+1), c)
}

// Fill tiles one sanitized grapheme across the grid using FillRect semantics.
// It returns false for invalid clusters without changing the grid.
func (g *GraphemeGrid) Fill(text string, style CellStyle) bool {
	return g.FillRect(g.Bounds(), text, style)
}

// FillLine tiles one sanitized grapheme across row y using FillRect semantics.
// It returns false for an invalid row or cluster without changing the grid.
func (g *GraphemeGrid) FillLine(y int, text string, style CellStyle) bool {
	if y < 0 || y >= g.Height() {
		return false
	}
	return g.FillRect(image.Rect(0, y, g.Width(), y+1), text, style)
}

// SetStyle replaces every cell style without changing content or the default
// Style used for future writes.
func (g *RuneGrid) SetStyle(s CellStyle) { g.SetStyleRect(g.Bounds(), s) }

// SetStyle replaces every cell style without changing content or the default
// Style used for future writes.
func (g *GraphemeGrid) SetStyle(s CellStyle) { g.SetStyleRect(g.Bounds(), s) }

// SetLines writes each string to a row starting at (0, 0), using the default
// Style and SetText semantics. Extra rows are clipped; unwritten cells remain.
func (g *RuneGrid) SetLines(lines []string) { g.SetLinesWithStyle(lines, g.Style) }

// SetLinesWithStyle is SetLines with an explicit style. It preserves cells
// beyond each written line and clips rows outside Bounds.
func (g *RuneGrid) SetLinesWithStyle(lines []string, style CellStyle) {
	for y, line := range lines {
		if y >= g.Height() {
			break
		}
		g.SetTextWithStyle(Point{Y: y}, line, style)
	}
}

// SetLines writes each string to a row starting at (0, 0), using the default
// Style and SetText semantics. Extra rows are clipped; unwritten cells remain.
func (g *GraphemeGrid) SetLines(lines []string) { g.SetLinesWithStyle(lines, g.Style) }

// SetLinesWithStyle is SetLines with an explicit style. It preserves cells
// beyond each written line and clips rows outside Bounds.
func (g *GraphemeGrid) SetLinesWithStyle(lines []string, style CellStyle) {
	for y, line := range lines {
		if y >= g.Height() {
			break
		}
		g.SetTextWithStyle(Point{Y: y}, line, style)
	}
}

// ShiftUp moves stored cells up one row and clears the bottom row.
func (g *RuneGrid) ShiftUp() { g.shiftUp() }

// ShiftDown moves stored cells down one row and clears the top row.
func (g *RuneGrid) ShiftDown() { g.shiftDown() }

// ShiftLeft moves stored cells left one column and clears the right edge.
func (g *RuneGrid) ShiftLeft() { g.shiftLeft() }

// ShiftRight moves stored cells right one column and clears the left edge.
func (g *RuneGrid) ShiftRight() { g.shiftRight() }

// ShiftUp moves stored cells up one row and clears the bottom row.
func (g *GraphemeGrid) ShiftUp() { g.shiftUp() }

// ShiftDown moves stored cells down one row and clears the top row.
func (g *GraphemeGrid) ShiftDown() { g.shiftDown() }

// ShiftLeft moves one display column; a cluster cut at the left edge is cleared.
func (g *GraphemeGrid) ShiftLeft() {
	g.shiftLeft()
	if g.Width() == 0 {
		return
	}
	for _, row := range g.content {
		if row[0].Continuation {
			row[0] = GraphemeCell{}
		}
	}
}

// ShiftRight moves one display column; a cluster cut at the right edge is cleared.
func (g *GraphemeGrid) ShiftRight() {
	g.shiftRight()
	if g.Width() == 0 {
		return
	}
	for _, row := range g.content {
		if row[len(row)-1].Width == 2 {
			row[len(row)-1] = GraphemeCell{}
		}
	}
}
