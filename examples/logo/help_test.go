package main

import (
	tea "charm.land/bubbletea/v2"
	"github.com/NimbleMarkets/go-booba/actions/bubblehelp"
	zone "github.com/lrstanley/bubblezone/v2"
	"nimbleterminal.dev/ntcanvas"
	"testing"
	"time"
)

func TestHelpAndInput(t *testing.T) {
	z := zone.New()
	defer z.Close()
	m := model{getExampleCanvas1(), getExampleCanvas2(z), z}
	m.c1.SetZoneManager(z)
	m.c1.Focus()
	snapshot, diagnostics, err := bubblehelp.Snapshot(m.helpKeys(), "demo", "canvas", 1, bubblehelp.Options{})
	if err != nil || len(diagnostics) != 0 || len(snapshot.Actions) < 6 {
		t.Fatalf("help: %v %v %+v", err, diagnostics, snapshot)
	}
	next, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	m = next.(model)
	if m.c1.Cursor().X != 1 || m.c2.Cursor().X != 0 {
		t.Fatal("focused viewport movement")
	}
	next, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	m = next.(model)
	next, _ = m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	m = next.(model)
	if m.c1.Focused() || !m.c2.Focused() || m.c2.Cursor().Y != 1 {
		t.Fatal("focus switch/wheel")
	}
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	if cmd == nil {
		t.Fatal("quit")
	}
}

func TestClickFocusAndDragOnlySelectedCanvas(t *testing.T) {
	z := zone.New()
	defer z.Close()
	m := model{getExampleCanvas1(), getExampleCanvas2(z), z}
	m.c1.SetZoneManager(z)
	m.c1.Focus()
	m.View()
	deadline := time.Now().Add(time.Second)
	for (z.Get(m.c1.ZoneID()).IsZero() || z.Get(m.c2.ZoneID()).IsZero()) && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	a, b := z.Get(m.c1.ZoneID()), z.Get(m.c2.ZoneID())
	if a.IsZero() || b.IsZero() {
		t.Fatal("zones unavailable")
	}
	send := func(msg tea.Msg) { next, _ := m.Update(msg); m = next.(model) }
	send(tea.MouseClickMsg{X: b.StartX + 1, Y: b.StartY + 1, Button: tea.MouseLeft})
	send(tea.MouseMotionMsg{X: b.StartX + 2, Y: b.StartY + 2, Button: tea.MouseLeft})
	if m.c1.Focused() || !m.c2.Focused() || m.c1.Cursor() != (ntcanvas.Point{}) || m.c2.Cursor() != (ntcanvas.Point{X: 1, Y: 1}) {
		t.Fatal(m.statusSummary())
	}
	send(tea.MouseClickMsg{X: a.StartX + 1, Y: a.StartY + 1, Button: tea.MouseLeft})
	send(tea.MouseMotionMsg{X: a.StartX + 2, Y: a.StartY + 2, Button: tea.MouseLeft})
	if !m.c1.Focused() || m.c2.Focused() || m.c1.Cursor() != (ntcanvas.Point{X: 1, Y: 1}) || m.c2.Cursor() != (ntcanvas.Point{X: 1, Y: 1}) {
		t.Fatal(m.statusSummary())
	}
	send(tea.MouseClickMsg{X: 100, Y: 100, Button: tea.MouseLeft})
	send(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	if m.c1.Focused() || m.c2.Focused() || m.c1.Cursor().Y != 1 || m.c2.Cursor().Y != 1 {
		t.Fatal("outside click should blur both", m.statusSummary())
	}
}
