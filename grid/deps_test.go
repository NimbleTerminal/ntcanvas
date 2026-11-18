package ntcanvasgrid

import (
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

// The point of this package is what it does not link: a program that only
// fills a grid and renders it must not pull in the terminal UI stack.
func TestGridLinksNoTerminalUI(t *testing.T) {
	if runtime.GOARCH == "wasm" {
		t.Skip("WebAssembly cannot run go list as a subprocess")
	}
	out, err := exec.Command("go", "list", "-deps", ".").CombinedOutput()
	if err != nil {
		t.Fatalf("go list -deps . failed: %v\n%s", err, out)
	}
	for _, dep := range strings.Fields(string(out)) {
		for _, banned := range []string{"charm.land/bubbletea", "charm.land/lipgloss", "charm.land/bubbles", "github.com/lrstanley/bubblezone", "github.com/charmbracelet/ultraviolet"} {
			if strings.HasPrefix(dep, banned) {
				t.Errorf("package ntcanvasgrid links %s", dep)
			}
		}
	}
}
