package ntcanvasgrid

import (
	"os/exec"
	"strings"
	"testing"
)

// The point of this package is what it does not link: a program that only
// fills a grid and renders it must not pull in the terminal UI stack.
func TestGridLinksNoTerminalUI(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", ".").Output()
	if err != nil {
		t.Skipf("go list is unavailable: %v", err)
	}
	for _, dep := range strings.Fields(string(out)) {
		for _, banned := range []string{"charm.land/bubbletea", "charm.land/lipgloss", "charm.land/bubbles", "github.com/lrstanley/bubblezone", "github.com/charmbracelet/ultraviolet"} {
			if strings.HasPrefix(dep, banned) {
				t.Errorf("package ntcanvasgrid links %s", dep)
			}
		}
	}
}
