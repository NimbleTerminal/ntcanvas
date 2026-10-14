// ntcharts - Copyright (c) 2024 Neomantra Corp.

package main

import (
	"charm.land/bubbles/v2/key"
	"fmt"
	"os"

	"nimbleterminal.dev/ntcanvas"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	zone "github.com/lrstanley/bubblezone/v2"
)

var highlightStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("1")). // red
	Background(lipgloss.Color("2"))  // green

var c1Style = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("63")) // purple

var c2Style = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("14")). // cyan
	Foreground(lipgloss.Color("212")).      // pink
	Background(lipgloss.Color("227"))       // yellow

type model struct {
	c1 ntcanvas.Model
	c2 ntcanvas.Model
	zM *zone.Manager
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.helpKeys().Toggle):
			if m.c1.Focused() {
				m.c1.Blur()
				m.c2.Focus()
			} else {
				m.c1.Focus()
				m.c2.Blur()
			}
		case key.Matches(msg, m.helpKeys().Quit):
			return m, tea.Quit
		}
	case tea.MouseClickMsg:
		if m.zM.Get(m.c1.ZoneID()).InBounds(msg) { // switch to canvas 1 if clicked on it
			m.c2.Blur()
			m.c1.Focus()
		} else if m.zM.Get(m.c2.ZoneID()).InBounds(msg) { // switch to canvas 2 if clicked on it
			m.c1.Blur()
			m.c2.Focus()
		} else {
			m.c1.Blur()
			m.c2.Blur()
		}
	}

	if m.c1.Focused() {
		m.c1, _ = m.c1.Update(msg)
	} else {
		m.c2, _ = m.c2.Update(msg)
	}

	return m, nil
}

func (m model) View() tea.View {
	s := "left click to select canvas, or `enter` to toggle between canvas\n"
	s += "click and drag, mouse wheel scroll or arrow keys to move viewport around\n"
	s += "`q/ctrl+c` to quit\n"
	s += lipgloss.JoinHorizontal(lipgloss.Top,
		c1Style.Render(m.c1.View()),
		c2Style.Render(m.c2.View())) + "\n"
	v := tea.NewView(m.zM.Scan(s)) // call zone Manager.Scan() at root model
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func getExampleCanvas1() (c ntcanvas.Model) {
	c1 := ntcanvas.NewStyledRuneGrid(20, 10)

	// set all contents at once with []string
	c1.SetLines([]string{
		"   ", // line 0 - missing characters will be displayed as ' ' up to width of Canvas
		" ██████████████████ ", // line 1
		" █                █ ", // line 2
		" █ ██████████████ █ ", // line 3
		" █ █                ", // line 4
		"                █ █ ", // line 5
		" █ ██████████████ █ ", // line 6
		" █                █ ", // line 7
		" ██████████████████ ", // line 8
		// line 9 - missing line will be displayed as ' ' up to width of Canvas
	})

	// Canvas coordinate system uses (0,0) as top left of Canvas
	// set runes in line 4 using string starting at (X,Y) coordinates (7,4) (\u2588 is █)
	c1.SetText(ntcanvas.Point{X: 7, Y: 4}, " IMBLE   \u2588 █ this will be dropped")

	// set runes in line 5 using []rune starting at (X,Y) coordinates (1,5)
	// (\u2588 and 0x2588 are both █)
	c1.SetRunes(ntcanvas.Point{X: 1, Y: 5}, []rune{'\u2588', ' ', 0x2588, ' ', ' ', ' ', 'T', 'E', 'R', 'M', 'I', 'N', 'A', 'L'})

	// set specific Cell at coordinates (7, 4)
	c1.SetCell(ntcanvas.Point{X: 7, Y: 4}, ntcanvas.NewCellWithStyle('N', highlightStyle))

	// set specific Cell styles at coordinates (7, 5)
	c1.SetCellStyle(ntcanvas.Point{X: 7, Y: 5}, highlightStyle) // 'T'

	return ntcanvas.NewModel(c1, ntcanvas.WithViewHeight(6), ntcanvas.WithViewWidth(18))
}

func getExampleCanvas2(zm *zone.Manager) (c ntcanvas.Model) {
	// set all contents at once with []string
	contents := []string{
		" ███████████████████  ",
		"  ███████████████████ ",
		"     ███       ███    ",
		"     ███       ███    ",
		"   ███████████████    ",
		"    ██████████████    ",
		"      ██       ███    ",
		"       █       ███    ",
		"                ██    ",
		"                 █    ",
	}

	// creating canvas with options
	grid := ntcanvas.NewStyledRuneGrid(30, 10, ntcanvas.WithContent(contents))
	c2 := ntcanvas.NewModel(grid,
		ntcanvas.WithZoneManager(zm),
		ntcanvas.WithViewHeight(6),
		ntcanvas.WithViewWidth(24))
	return c2
}

func main() {
	// Mouse support is enabled using the bubblezone module
	// 1. create a new bubblezone.Manager with bubblezone.New()
	// 2. call ntcanvas.SetZoneManager() on the new bubblezone.Manager
	// 3. create bubbletea.Program with mouse support enabled, e.g. tea.WithMouseCellMotion()
	// 4. bubblezone.Manager.Scan() must be called in bubbletea Model.View()
	//    wrapping the returned string output
	z := zone.New()
	defer z.Close()
	m := model{getExampleCanvas1(), getExampleCanvas2(z), z}
	m.c1.SetZoneManager(z)
	m.c1.Focus()

	// booba.Run is a tea.Program substitute that dispatches to native Bubble Tea
	// or the WASM/ghostty-web bridge depending on build target.
	// See https://github.com/NimbleMarkets/go-booba-example.
	if err := runDemo(m); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
