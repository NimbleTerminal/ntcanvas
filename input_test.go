package ntcanvas

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	zone "github.com/lrstanley/bubblezone/v2"
	"testing"
	"time"
)

func TestInputDirectionsAndBounds(t *testing.T) {
	cases := []struct {
		name string
		msg  tea.Msg
		want Point
	}{
		{"up", tea.KeyPressMsg{Code: tea.KeyUp}, Point{X: 2, Y: 1}},
		{"down", tea.KeyPressMsg{Code: tea.KeyDown}, Point{X: 2, Y: 3}},
		{"left", tea.KeyPressMsg{Code: tea.KeyLeft}, Point{X: 1, Y: 2}},
		{"right", tea.KeyPressMsg{Code: tea.KeyRight}, Point{X: 3, Y: 2}},
		{"pgup", tea.KeyPressMsg{Code: tea.KeyPgUp}, Point{X: 2, Y: 1}},
		{"pgdown", tea.KeyPressMsg{Code: tea.KeyPgDown}, Point{X: 2, Y: 3}},
		{"wheel up", tea.MouseWheelMsg{Button: tea.MouseWheelUp}, Point{X: 2, Y: 1}},
		{"wheel down", tea.MouseWheelMsg{Button: tea.MouseWheelDown}, Point{X: 2, Y: 3}},
		{"wheel left", tea.MouseWheelMsg{Button: tea.MouseWheelLeft}, Point{X: 1, Y: 2}},
		{"wheel right", tea.MouseWheelMsg{Button: tea.MouseWheelRight}, Point{X: 3, Y: 2}},
		{"unknown", tea.KeyPressMsg{Code: 'x', Text: "x"}, Point{X: 2, Y: 2}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := NewModel(NewStyledRuneGrid(6, 6), WithFocus(), WithViewWidth(3), WithViewHeight(3), WithCursor(Point{X: 2, Y: 2}))
			m, cmd := m.Update(tc.msg)
			if cmd != nil || m.Cursor() != tc.want {
				t.Fatalf("cursor %v, want %v", m.Cursor(), tc.want)
			}
			m.Blur()
			m, _ = m.Update(tc.msg)
			if m.Cursor() != tc.want {
				t.Fatal("blurred input moved viewport")
			}
		})
	}
	m := NewModel(NewStyledRuneGrid(4, 4), WithFocus(), WithViewWidth(2), WithViewHeight(2))
	for range 8 {
		m.MoveRight(1)
		m.MoveDown(1)
	}
	if m.Cursor() != (Point{X: 2, Y: 2}) {
		t.Fatal("bottom/right bound", m.Cursor())
	}
	for range 8 {
		m.MoveUp(1)
		m.MoveLeft(1)
	}
	if m.Cursor() != (Point{}) {
		t.Fatal("top/left bound", m.Cursor())
	}
	k := DefaultKeyMap()
	k.Right = key.NewBinding(key.WithKeys("x"))
	m = NewModel(NewStyledRuneGrid(4, 4), WithFocus(), WithViewWidth(2), WithKeyMap(k))
	m, _ = m.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	if m.Cursor().X != 1 {
		t.Fatal("custom key map ignored")
	}
	m.KeyMap.Right.SetEnabled(false)
	m, _ = m.Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	if m.Cursor().X != 1 {
		t.Fatal("disabled binding moved")
	}
}

func TestKeyReleaseDoesNotPan(t *testing.T) {
	for _, code := range []rune{tea.KeyUp, tea.KeyDown, tea.KeyLeft, tea.KeyRight, tea.KeyPgUp, tea.KeyPgDown, 'h', 'j', 'k', 'l'} {
		m := NewModel(NewRuneGrid(10, 10), WithFocus(), WithViewWidth(2), WithViewHeight(2), WithCursor(Point{X: 4, Y: 4}))
		m, _ = m.Update(tea.KeyPressMsg{Code: code})
		want := m.Cursor()
		if want == (Point{X: 4, Y: 4}) {
			t.Fatalf("press %d did not pan", code)
		}
		m, _ = m.Update(tea.KeyReleaseMsg{Code: code})
		if m.Cursor() != want {
			t.Fatalf("release %d moved cursor from %v to %v", code, want, m.Cursor())
		}
	}
}

func TestZoneClickAndDrag(t *testing.T) {
	z := zone.New()
	defer z.Close()
	m := NewModel(NewStyledRuneGrid(6, 6), WithFocus(), WithZoneManager(z), WithViewWidth(3), WithViewHeight(3))
	z.Scan(m.View())
	deadline := time.Now().Add(time.Second)
	for z.Get(m.ZoneID()).IsZero() && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if z.Get(m.ZoneID()).IsZero() {
		t.Fatal("zone scan did not complete")
	}
	if m.ZoneManager() != z {
		t.Fatal("zone manager not retained")
	}
	click := tea.MouseClickMsg{X: 1, Y: 1, Button: tea.MouseLeft}
	m, _ = m.Update(click)
	if m.Cursor() != (Point{}) {
		t.Fatal("click alone moved")
	}
	m, _ = m.Update(tea.MouseMotionMsg{X: 2, Y: 2, Button: tea.MouseLeft})
	if m.Cursor() != (Point{X: 1, Y: 1}) {
		t.Fatal("drag down/right", m.Cursor())
	}
	m, _ = m.Update(tea.MouseMotionMsg{X: 100, Y: 100, Button: tea.MouseLeft})
	m, _ = m.Update(tea.MouseMotionMsg{X: 1, Y: 1, Button: tea.MouseNone})
	if m.Cursor() != (Point{X: 1, Y: 1}) {
		t.Fatal("outside drag or hover moved")
	}
	m, _ = m.Update(tea.MouseMotionMsg{X: 1, Y: 1, Button: tea.MouseLeft})
	if m.Cursor() != (Point{}) {
		t.Fatal("drag up/left", m.Cursor())
	}
	// Two model copies share a zone/surface but must not share drag history.
	a, b := m, m
	a, _ = a.Update(tea.MouseClickMsg{X: 0, Y: 0, Button: tea.MouseLeft})
	b, _ = b.Update(tea.MouseClickMsg{X: 2, Y: 2, Button: tea.MouseLeft})
	a, _ = a.Update(tea.MouseMotionMsg{X: 1, Y: 1, Button: tea.MouseLeft})
	if a.Cursor() != (Point{X: 1, Y: 1}) || b.Cursor() != (Point{}) {
		t.Fatal("shared drag closure")
	}
	a, _ = a.Update(tea.MouseReleaseMsg{X: 1, Y: 1, Button: tea.MouseLeft})
	a, _ = a.Update(tea.MouseMotionMsg{X: 2, Y: 2, Button: tea.MouseLeft})
	if a.Cursor() != (Point{X: 1, Y: 1}) {
		t.Fatal("drag after release")
	}
	m.SetZoneManager(nil)
	m, _ = m.Update(click)
	m, _ = m.Update(tea.MouseMotionMsg{X: 2, Y: 2, Button: tea.MouseLeft})
	if m.Cursor() != (Point{}) || m.ZoneManager() != nil {
		t.Fatal("disabled zone moved")
	}
}

func TestModelKeyMapOwnership(t *testing.T) {
	a, b := NewModel(nil), NewModel(nil)
	if a.KeyMap == b.KeyMap || a.KeyMap == nil {
		t.Fatal("constructors share key map")
	}
	a.KeyMap.Right.SetEnabled(false)
	if !b.KeyMap.Right.Enabled() {
		t.Fatal("defaults leaked")
	}
	keys := DefaultKeyMap()
	option := WithKeyMap(keys)
	a, b = NewModel(nil, option), NewModel(nil, option)
	a.KeyMap.Right.SetEnabled(false)
	if !keys.Right.Enabled() || !b.KeyMap.Right.Enabled() {
		t.Fatal("option reused mutable key map")
	}
	copy := b
	copy.KeyMap.Right.SetEnabled(false)
	if b.KeyMap.Right.Enabled() {
		t.Fatal("model copies should share bindings")
	}
	detached := *copy.KeyMap
	copy.KeyMap = &detached
	copy.KeyMap.Right.SetEnabled(true)
	if b.KeyMap.Right.Enabled() {
		t.Fatal("detached key map still shared")
	}
	m := NewModel(NewRuneGrid(4, 4), WithFocus(), WithViewWidth(2), WithViewHeight(2))
	m.KeyMap = nil
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
	if m.Cursor() != (Point{}) {
		t.Fatal("nil key map moved viewport")
	}
	m, _ = m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	if m.Cursor() != (Point{Y: 1}) {
		t.Fatal("nil key map disabled mouse")
	}
}
