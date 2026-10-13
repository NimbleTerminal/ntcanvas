package ntcanvas

import "image"

func (g *RuneGrid) Width() int      { return g.area.Dx() }
func (g *RuneGrid) Height() int     { return g.area.Dy() }
func (g *GraphemeGrid) Width() int  { return g.area.Dx() }
func (g *GraphemeGrid) Height() int { return g.area.Dy() }

func (g *RuneGrid) Fill(c RuneCell) { g.FillRect(g.Bounds(), c) }
func (g *RuneGrid) FillLine(y int, c RuneCell) {
	if y < 0 || y >= g.Height() {
		return
	}
	g.FillRect(image.Rect(0, y, g.Width(), y+1), c)
}
func (g *GraphemeGrid) Fill(text string, style CellStyle) bool {
	return g.FillRect(g.Bounds(), text, style)
}
func (g *GraphemeGrid) FillLine(y int, text string, style CellStyle) bool {
	if y < 0 || y >= g.Height() {
		return false
	}
	return g.FillRect(image.Rect(0, y, g.Width(), y+1), text, style)
}
func (g *RuneGrid) SetStyle(s CellStyle)     { g.SetStyleRect(g.Bounds(), s) }
func (g *GraphemeGrid) SetStyle(s CellStyle) { g.SetStyleRect(g.Bounds(), s) }

func (g *RuneGrid) SetLines(lines []string) { g.SetLinesWithStyle(lines, g.Style) }
func (g *RuneGrid) SetLinesWithStyle(lines []string, style CellStyle) {
	for y, line := range lines {
		if y >= g.Height() {
			break
		}
		g.SetTextWithStyle(Point{Y: y}, line, style)
	}
}
func (g *GraphemeGrid) SetLines(lines []string) { g.SetLinesWithStyle(lines, g.Style) }
func (g *GraphemeGrid) SetLinesWithStyle(lines []string, style CellStyle) {
	for y, line := range lines {
		if y >= g.Height() {
			break
		}
		g.SetTextWithStyle(Point{Y: y}, line, style)
	}
}

// Shifts move one stored column or row and clear the exposed edge.
func (g *RuneGrid) ShiftUp()       { g.shiftUp() }
func (g *RuneGrid) ShiftDown()     { g.shiftDown() }
func (g *RuneGrid) ShiftLeft()     { g.shiftLeft() }
func (g *RuneGrid) ShiftRight()    { g.shiftRight() }
func (g *GraphemeGrid) ShiftUp()   { g.shiftUp() }
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
