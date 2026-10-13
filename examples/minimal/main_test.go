package main

import (
	tea "charm.land/bubbletea/v2"
	"nimbleterminal.dev/ntcanvas"
	"testing"
)

func TestHostRetainsUpdateAndHandlesTinyWindows(t *testing.T) {
	m := newModel()
	next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	m = next.(model)
	if m.canvas.Cursor() != (ntcanvas.Point{X: 1}) {
		t.Fatal("host discarded canvas update")
	}
	for _, size := range []tea.WindowSizeMsg{{Width: 1, Height: 1}, {Width: 0, Height: 0}, {Width: -1, Height: -1}} {
		next, _ = m.Update(size)
		m = next.(model)
		if !m.canvas.Viewport().Empty() {
			t.Fatal("tiny terminal viewport")
		}
		m.View()
	}
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	if cmd == nil {
		t.Fatal("host quit")
	}
}
