// Package demohelp connects example help keys to native and browser controls.
package demohelp

import (
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	booba "github.com/NimbleMarkets/go-booba"
	"github.com/NimbleMarkets/go-booba/actions"
	"github.com/NimbleMarkets/go-booba/actions/bubblehelp"
	"github.com/charmbracelet/x/ansi"
)

type Demo interface {
	tea.Model
	Help() []key.Binding
	Summary() string
}
type helpKeys []key.Binding

func (k helpKeys) ShortHelp() []key.Binding  { return k }
func (k helpKeys) FullHelp() [][]key.Binding { return [][]key.Binding{k} }
func Key(label string, keys ...string) key.Binding {
	return key.NewBinding(key.WithKeys(keys...), key.WithHelp(keys[0], label))
}
func Quit(msg tea.KeyPressMsg) bool { return msg.String() == "q" || msg.String() == "ctrl+c" }
func Main(m Demo) {
	wrapped, err := bubblehelp.NewModel(m, func(inner tea.Model) bubblehelp.Description {
		d := inner.(Demo)
		return bubblehelp.Description{Context: "ntcanvas-demo", Keys: helpKeys(d.Help()), Status: actions.ProgramStatus{App: "ntcanvas", Summary: d.Summary()}}
	}, bubblehelp.ModelOptions{Adapter: bubblehelp.Options{Customize: func(_ key.Binding, a *actions.Action) error { first := 0; a.Preferred = &first; return nil }}})
	if err == nil {
		err = booba.Run(wrapped)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// Fit bounds root output on small terminals. Demos cap dimensions before arithmetic.
func Fit(text string, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	lines := strings.Split(text, "\n")
	lines = lines[:min(len(lines), height)]
	for i := range lines {
		lines[i] = ansi.Truncate(lines[i], width, "")
	}
	return strings.Join(lines, "\n")
}
func View(text string, width, height int) tea.View {
	v := tea.NewView(Fit(text, width, height))
	v.AltScreen = true
	return v
}
