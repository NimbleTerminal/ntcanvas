// Compare connected cell lines with caller-mapped antialias coverage.
package main

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"fmt"
	"math"
	"nimbleterminal.dev/ntcanvas"
	"nimbleterminal.dev/ntcanvas/examples/internal/demohelp"
)

type model struct {
	step          int
	blocks        bool
	width, height int
}

func newModel() model         { return model{width: 80, height: 28} }
func (m model) Init() tea.Cmd { return nil }
func (m model) Help() []key.Binding {
	return []key.Binding{demohelp.Key("Raise endpoint", "up", "k"), demohelp.Key("Lower endpoint", "down", "j"), demohelp.Key("Toggle glyph", "g"), demohelp.Key("Reset", "r"), demohelp.Key("Quit", "q", "ctrl+c")}
}
func (m model) Summary() string { return fmt.Sprintf("Lines: offset=%d blocks=%t", m.step, m.blocks) }
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = min(160, max(0, msg.Width)), min(60, max(0, msg.Height))
	case tea.KeyPressMsg:
		if demohelp.Quit(msg) {
			return m, tea.Quit
		}
		switch msg.String() {
		case "up", "k":
			m.step = max(-12, m.step-1)
		case "down", "j":
			m.step = min(12, m.step+1)
		case "g":
			m.blocks = !m.blocks
		case "r":
			m.step = 0
			m.blocks = false
		}
	}
	return m, nil
}
func lineGrid(w, h, step int, blocks, aa bool) *ntcanvas.RuneGrid {
	g := ntcanvas.NewRuneGrid(w, h)
	bg := ntcanvas.RGBColor(16, 24, 32)
	g.Fill(ntcanvas.RuneCell{Rune: ' ', Style: ntcanvas.CellStyle{Background: bg}})
	a, b := ntcanvas.Float64Point{X: -6, Y: 1.25}, ntcanvas.Float64Point{X: float64(w) + 5, Y: float64(h) - 2.25 + float64(step)/2}
	glyph := '·'
	if blocks {
		glyph = '█'
	}
	plot := func(p ntcanvas.Point, coverage float64) {
		// Blend against the same explicit background used to fill the grid.
		blend := func(back, front float64) uint8 { return uint8(math.Round(back + (front-back)*coverage)) }
		g.SetRuneWithStyle(p, glyph, ntcanvas.CellStyle{Foreground: ntcanvas.RGBColor(blend(16, 141), blend(24, 227), blend(32, 182)), Background: bg})
	}
	if aa {
		ntcanvas.RasterLineAA(g.Bounds(), a, b, plot)
	} else {
		ntcanvas.RasterLine(g.Bounds(), a, b, func(p ntcanvas.Point) { plot(p, 1) })
	}
	return g
}
func (m model) View() tea.View {
	w, h := min(30, max(1, (m.width-6)/2)), min(12, max(2, m.height-9))
	narrow := m.width < 64
	if narrow {
		w = min(40, max(1, m.width-2))
		h = min(8, max(2, (m.height-10)/2))
	}
	panel := func(aa bool) string {
		name := "RasterLine"
		if aa {
			name = "RasterLineAA"
		}
		g := lineGrid(w, h, m.step, m.blocks, aa)
		return name + "\n" + lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Render(g.Render(g.Bounds()))
	}
	body := lipgloss.JoinHorizontal(lipgloss.Top, panel(false), "  ", panel(true))
	if narrow {
		body = panel(false) + "\n" + panel(true)
	}
	return demohelp.View("CLIPPED LINES\n↑/↓: endpoint · g: glyph · r: reset · q: quit\n"+m.Summary()+"\n\n"+body+"\n\nAA coverage blends foreground against an explicit background.\nBoth segments start and end outside the grid.", m.width, m.height)
}
func main() { demohelp.Main(newModel()) }
