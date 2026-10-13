// Two independent Models view and edit one shared GraphemeGrid.
package main

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"fmt"
	zone "github.com/lrstanley/bubblezone/v2"
	"nimbleterminal.dev/ntcanvas"
	"nimbleterminal.dev/ntcanvas/examples/internal/demohelp"
)

type model struct {
	grid                 *ntcanvas.GraphemeGrid
	left, right          ntcanvas.Model
	zones                *zone.Manager
	width, height, marks int
	small                bool
}

func newModel(z *zone.Manager) model {
	g := ntcanvas.NewGraphemeGrid(72, 18)
	m := model{grid: g, zones: z, width: 80, height: 28}
	m.left = ntcanvas.NewModel(g, ntcanvas.WithFocus(), ntcanvas.WithZoneManager(z))
	m.right = ntcanvas.NewModel(g, ntcanvas.WithCursor(ntcanvas.Point{X: 12, Y: 4}), ntcanvas.WithZoneManager(z))
	m.paint()
	m.layout()
	return m
}
func (m *model) paint() {
	m.grid.Clear()
	for y := 0; y < m.grid.Height(); y++ {
		for x := 0; x < m.grid.Width(); x++ {
			ch := "·"
			if x%10 == 0 {
				ch = fmt.Sprint(x / 10)
			}
			m.grid.SetGrapheme(ntcanvas.Point{X: x, Y: y}, ch, ntcanvas.CellStyle{Foreground: ntcanvas.IndexedColor(8)})
		}
		m.grid.SetTextWithStyle(ntcanvas.Point{Y: y}, fmt.Sprintf("%02d", y), ntcanvas.CellStyle{Foreground: ntcanvas.RGBColor(141, 227, 182)})
	}
	m.grid.SetText(ntcanvas.Point{X: 14, Y: 4}, "Shared surface 界")
	if m.marks > 0 {
		m.grid.SetTextWithStyle(ntcanvas.Point{X: 20, Y: 6}, fmt.Sprintf("EDIT %d", m.marks), ntcanvas.CellStyle{Foreground: ntcanvas.RGBColor(255, 190, 100), Attributes: ntcanvas.Bold})
	}
}
func (m *model) layout() {
	w, h := min(32, max(1, (m.width-6)/2)), min(10, max(1, m.height-9))
	if m.width < 64 {
		w = min(40, max(1, m.width-2))
		h = min(6, max(1, (m.height-10)/2))
	}
	// Resize view windows without changing storage or resetting their cursors.
	m.left.ViewWidth, m.right.ViewWidth = w, w
	m.left.ViewHeight, m.right.ViewHeight = h, h
}
func (m model) Init() tea.Cmd { return nil }
func (m model) Help() []key.Binding {
	return []key.Binding{
		demohelp.Key("Move up", "up", "k"), demohelp.Key("Move down", "down", "j"), demohelp.Key("Move left", "left", "h"), demohelp.Key("Move right", "right", "l"),
		demohelp.Key("Switch viewport", "enter"), demohelp.Key("Edit shared grid", "e"), demohelp.Key("Resize grid", "g"), demohelp.Key("Blur", "b"), demohelp.Key("Reset views", "r"), demohelp.Key("Quit", "q", "ctrl+c"),
	}
}
func (m model) Summary() string {
	focus := "none"
	if m.left.Focused() {
		focus = "left"
	} else if m.right.Focused() {
		focus = "right"
	}
	a, b := m.left.Cursor(), m.right.Cursor()
	return fmt.Sprintf("Viewports: focus=%s left=(%d,%d) right=(%d,%d) grid=%dx%d marks=%d", focus, a.X, a.Y, b.X, b.Y, m.grid.Width(), m.grid.Height(), m.marks)
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = min(160, max(0, msg.Width)), min(60, max(0, msg.Height))
		m.layout()
	case tea.KeyPressMsg:
		if demohelp.Quit(msg) {
			return m, tea.Quit
		}
		switch msg.String() {
		case "enter":
			if m.left.Focused() {
				m.left.Blur()
				m.right.Focus()
			} else {
				m.right.Blur()
				m.left.Focus()
			}
		case "b":
			m.left.Blur()
			m.right.Blur()
		case "e":
			m.marks++
			m.paint()
		case "g":
			m.small = !m.small
			w, h := 72, 18
			if m.small {
				w, h = 28, 10
			}
			// These dimensions are application constants. External dimensions should use ResizeChecked.
			m.grid.Resize(w, h)
			m.paint()
		case "r":
			m.left.SetCursor(ntcanvas.Point{})
			m.right.SetCursor(ntcanvas.Point{X: 12, Y: 4})
			m.left.Focus()
			m.right.Blur()
		}
	case tea.MouseClickMsg:
		m.left.Blur()
		m.right.Blur()
		if m.zones.Get(m.left.ZoneID()).InBounds(msg) {
			m.left.Focus()
		} else if m.zones.Get(m.right.ZoneID()).InBounds(msg) {
			m.right.Focus()
		}
	}
	m.left, _ = m.left.Update(msg)
	m.right, _ = m.right.Update(msg)
	return m, nil
}
func (m model) View() tea.View {
	panel := func(name string, c ntcanvas.Model) string {
		color := "8"
		if c.Focused() {
			color = "14"
			name += " · focused"
		}
		p := c.Cursor()
		return fmt.Sprintf("%s (%d,%d)\n", name, p.X, p.Y) + lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color(color)).Render(c.View())
	}
	left, right := panel("Left", m.left), panel("Right", m.right)
	body := lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", right)
	if m.width < 64 {
		body = left + "\n" + right
	}
	text := "ONE GRID, TWO VIEWPORTS\nEnter: focus · arrows/HJKL, wheel or drag: pan\ne: shared edit · g: grid size · b: blur · r: reset · q: quit\n\n" + body + "\n\n" + m.Summary()
	v := tea.NewView(m.zones.Scan(demohelp.Fit(text, m.width, m.height)))
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}
func main() { z := zone.New(); defer z.Close(); demohelp.Main(newModel(z)) }
