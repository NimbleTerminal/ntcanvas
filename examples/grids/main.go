// Compare cell storage and styling without changing the same input text.
package main

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"fmt"
	"image"
	"nimbleterminal.dev/ntcanvas"
	"nimbleterminal.dev/ntcanvas/examples/internal/demohelp"
	"strings"
)

type model struct {
	styled        *ntcanvas.StyledRuneGrid
	runes         *ntcanvas.RuneGrid
	graphemes     *ntcanvas.GraphemeGrid
	safe, cleared bool
	width, height int
}

func newModel() model { m := model{width: 80, height: 28}; m.reset(); return m }
func (m *model) reset() {
	m.styled = ntcanvas.NewStyledRuneGrid(44, 3)
	m.runes = ntcanvas.NewRuneGrid(44, 3)
	m.graphemes = ntcanvas.NewGraphemeGrid(44, 3)
	lines := []string{"Every cell has its own style.", "Wide: 界語  marks: e\u0301  emoji: 👩‍💻", "0123456789  ← cell coordinates"}
	m.styled.SetLines(lines)
	m.runes.SetLines(lines)
	m.graphemes.SetLines(lines)
	area := image.Rect(0, 0, 10, 1)
	m.styled.SetStyleRect(area, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#8de3b6")))
	style := ntcanvas.CellStyle{Foreground: ntcanvas.RGBColor(141, 227, 182), Attributes: ntcanvas.Bold}
	m.runes.SetStyleRect(area, style)
	m.graphemes.SetStyleRect(area, style)
	m.safe, m.cleared = false, false
}
func (m model) Init() tea.Cmd { return nil }
func (m model) Help() []key.Binding {
	return []key.Binding{demohelp.Key("Toggle single column", "s"), demohelp.Key("Clear wide cell", "c"), demohelp.Key("Reset", "r"), demohelp.Key("Quit", "q", "ctrl+c")}
}
func (m model) Summary() string {
	return fmt.Sprintf("Grids: single-column=%t cleared=%t", m.safe, m.cleared)
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = min(160, max(0, msg.Width)), min(60, max(0, msg.Height))
	case tea.KeyPressMsg:
		if demohelp.Quit(msg) {
			return m, tea.Quit
		}
		switch msg.String() {
		case "s":
			m.safe = !m.safe
			m.styled.SingleColumn = m.safe
			m.runes.SingleColumn = m.safe
		case "c":
			// Column seven is the second half of 界 in GraphemeGrid. Clearing it
			// clears that entire grapheme; rune grids instead clear one stored rune.
			area := image.Rect(7, 1, 8, 2)
			m.styled.ClearRect(area)
			m.runes.ClearRect(area)
			m.graphemes.ClearRect(area)
			m.cleared = true
		case "r":
			m.reset()
		}
	}
	return m, nil
}
func (m model) View() tea.View {
	var b strings.Builder
	b.WriteString("GRID COMPARISON\ns: single-column · c: clear cell · r: reset · q: quit\n")
	b.WriteString(m.Summary() + "\n\n")
	for _, item := range []struct {
		name    string
		surface ntcanvas.Surface
	}{
		{"StyledRuneGrid · Lip Gloss per cell", m.styled},
		{"RuneGrid · compact terminal styles", m.runes},
		{"GraphemeGrid · display-column layout", m.graphemes},
	} {
		view := ntcanvas.NewModel(item.surface, ntcanvas.WithViewWidth(min(44, max(0, m.width-2))))
		b.WriteString(item.name + "\n" + view.View() + "\n\n")
	}
	b.WriteString("Rune grids advance once per rune. Graphemes keep wide glyphs together.\nSingleColumn replaces wide/combining runes; it does not reshape stored text.")
	return demohelp.View(b.String(), m.width, m.height)
}
func main() { demohelp.Main(newModel()) }
