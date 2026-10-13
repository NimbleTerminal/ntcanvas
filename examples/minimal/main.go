// Minimal Bubble Tea v2 integration: the host owns Update, View, and terminal size.
package main

import (
	tea "charm.land/bubbletea/v2"
	"fmt"
	"image"
	"nimbleterminal.dev/ntcanvas"
	"os"
)

type model struct{ canvas ntcanvas.Model }

func newModel() model {
	g, err := ntcanvas.NewGraphemeGridChecked(80, 24, ntcanvas.SizeLimits{})
	if err != nil {
		panic(err)
	} // fixed, trusted dimensions
	g.SetTextWithStyle(ntcanvas.Point{X: 2, Y: 1}, "Hello 界 · e\u0301 · 👩‍💻", ntcanvas.CellStyle{Attributes: ntcanvas.Bold})
	for y := 3; y < 24; y++ {
		g.SetText(ntcanvas.Point{X: 2, Y: y}, fmt.Sprintf("Row %02d — arrows or HJKL pan this canvas", y))
	}
	c := ntcanvas.NewModel(g, ntcanvas.WithFocus())
	c.SetViewport(image.Rect(0, 0, 40, 8))
	return model{canvas: c}
}
func (m model) Init() tea.Cmd { return m.canvas.Init() }
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		// Clamp before arithmetic; SetViewport also clips to the backing grid.
		m.canvas.SetViewport(image.Rect(0, 0, min(80, max(0, msg.Width)), max(0, min(26, msg.Height)-2)))
	}
	var cmd tea.Cmd
	m.canvas, cmd = m.canvas.Update(msg) // retain the returned viewport/focus state
	return m, cmd
}
func (m model) View() tea.View {
	view := tea.NewView(m.canvas.View() + "\nArrows/HJKL: pan • q: quit")
	view.AltScreen = true
	return view
}
func main() {
	if _, err := tea.NewProgram(newModel()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
