package main

import (
	tea "charm.land/bubbletea/v2"
	zone "github.com/lrstanley/bubblezone/v2"
	"nimbleterminal.dev/ntcanvas"
	"strings"
	"testing"
	"time"
)

func TestSharedSurfaceAndFocus(t *testing.T) {
	z := zone.New()
	defer z.Close()
	m := newModel(z)
	send := func(msg tea.Msg) { n, _ := m.Update(msg); m = n.(model) }
	press := func(k string) { send(tea.KeyPressMsg{Code: rune(k[0]), Text: k}) }
	press("e")
	if !strings.Contains(m.left.View(), "EDIT 1") || !strings.Contains(m.right.View(), "EDIT 1") {
		t.Fatal("edit was not shared")
	}
	if m.left.Surface() != m.right.Surface() {
		t.Fatal("separate backing grids")
	}
	send(tea.KeyPressMsg{Code: tea.KeyRight})
	if m.left.Cursor().X != 1 || m.right.Cursor().X != 12 {
		t.Fatal(m.Summary())
	}
	press("b")
	send(tea.KeyPressMsg{Code: tea.KeyRight})
	if m.left.Focused() || m.right.Focused() || m.left.Cursor().X != 1 {
		t.Fatal("blur", m.Summary())
	}
	m.right.SetCursor(ntcanvas.Point{X: 70, Y: 17})
	press("g")
	if m.right.Cursor() != (ntcanvas.Point{X: 27, Y: 9}) || m.left.Cursor().X != 1 {
		t.Fatal("resize clamp", m.Summary())
	}
	press("r")
	if !m.left.Focused() || m.right.Focused() {
		t.Fatal("reset focus")
	}
	for _, size := range []int{0, 1, 12, 40, 80, int(^uint(0) >> 1)} {
		send(tea.WindowSizeMsg{Width: size, Height: size})
		m.View()
	}
}
func TestViewportZoneMouse(t *testing.T) {
	z := zone.New()
	defer z.Close()
	m := newModel(z)
	m.View()
	deadline := time.Now().Add(time.Second)
	for z.Get(m.right.ZoneID()).IsZero() && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	info := z.Get(m.right.ZoneID())
	if info.IsZero() {
		t.Fatal("zone unavailable")
	}
	send := func(msg tea.Msg) { n, _ := m.Update(msg); m = n.(model) }
	send(tea.MouseClickMsg{X: info.StartX + 1, Y: info.StartY + 1, Button: tea.MouseLeft})
	send(tea.MouseMotionMsg{X: info.StartX + 2, Y: info.StartY + 2, Button: tea.MouseLeft})
	if m.left.Focused() || !m.right.Focused() || m.left.Cursor() != (ntcanvas.Point{}) || m.right.Cursor() != (ntcanvas.Point{X: 13, Y: 5}) {
		t.Fatal(m.Summary())
	}
	send(tea.MouseReleaseMsg{Button: tea.MouseLeft})
	send(tea.MouseClickMsg{X: 159, Y: 59, Button: tea.MouseLeft})
	send(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	if m.right.Focused() || m.right.Cursor().Y != 5 {
		t.Fatal("outside click", m.Summary())
	}
}
