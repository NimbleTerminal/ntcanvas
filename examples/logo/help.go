package main

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"fmt"
	booba "github.com/NimbleMarkets/go-booba"
	"github.com/NimbleMarkets/go-booba/actions"
	"github.com/NimbleMarkets/go-booba/actions/bubblehelp"
	"nimbleterminal.dev/ntcanvas"
)

type helpKeys struct {
	ntcanvas.KeyMap
	Toggle, Quit key.Binding
}

func (m model) helpKeys() helpKeys {
	k := m.c1.KeyMap
	if m.c2.Focused() {
		k = m.c2.KeyMap
	}
	keys := ntcanvas.KeyMap{}
	if k != nil {
		keys = *k
	}
	return helpKeys{keys,
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "Switch canvas")),
		key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "Quit"))}
}
func (k helpKeys) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Left, k.Right, k.Toggle}
}
func (k helpKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp(), {k.PgUp, k.PgDown, k.Quit}}
}
func runDemo(m model) error {
	wrapped, err := bubblehelp.NewModel(m, func(inner tea.Model) bubblehelp.Description {
		return bubblehelp.Description{Context: "canvas-logo", Keys: inner.(model).helpKeys(), Status: actions.ProgramStatus{App: "ntcanvas", Summary: inner.(model).statusSummary()}}
	}, bubblehelp.ModelOptions{Adapter: bubblehelp.Options{Customize: func(_ key.Binding, a *actions.Action) error { first := 0; a.Preferred = &first; return nil }}})
	if err != nil {
		return err
	}
	return booba.Run(wrapped)
}

// Expose the focused viewport and pan positions to booba's status UI.
func (m model) statusSummary() string {
	focus := "none"
	if m.c1.Focused() {
		focus = "left"
	} else if m.c2.Focused() {
		focus = "right"
	}
	a, b := m.c1.Cursor(), m.c2.Cursor()
	return fmt.Sprintf("Canvas viewports: focus=%s left=(%d,%d) right=(%d,%d)", focus, a.X, a.Y, b.X, b.Y)
}
