package ntcanvas

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"testing"
)

func TestRuneCellsAndClipping(t *testing.T) {
	m := NewStyledRuneGrid(4, 1)
	s := lipgloss.NewStyle().Bold(true)
	m.SetTextWithStyle(Point{}, "界e\u0301xZ", s)
	for x, r := range []rune("界e\u0301x") {
		if c := m.Cell(Point{X: x}); c.Rune != r || !c.Style.GetBold() {
			t.Fatalf("cell %d: %+v", x, c)
		}
	}
	want := ""
	for _, r := range "界e\u0301x" {
		want += s.Render(string(r))
	}
	if m.Render(m.Bounds()) != want {
		t.Fatalf("view %q != %q", m.Render(m.Bounds()), want)
	}
	m.SetText(Point{X: -2}, "abcd")
	if m.Cell(Point{}).Rune != 'c' {
		t.Fatal("left clipping")
	}
	if m.SetRune(Point{X: 4}, 'x') || m.SetText(Point{Y: 1}, "x") {
		t.Fatal("bounds")
	}
}
func TestResizeAndViewport(t *testing.T) {
	g := NewStyledRuneGrid(4, 3, WithContent([]string{"abcd", "efgh", "ijkl"}))
	m := NewModel(g, WithViewWidth(2), WithViewHeight(2), WithFocus())
	m.MoveRight(1)
	m.MoveDown(1)
	if m.View() != "fg\njk" {
		t.Fatal(m.View())
	}
	g.Resize(2, 2)
	if m.Cursor() != (Point{X: 1, Y: 1}) || m.View() != "f" {
		t.Fatal(m.Cursor(), m.View())
	}
	m.SetCursor(Point{})
	g.Resize(3, 3)
	if g.Cell(Point{X: 2, Y: 2}).Rune != 0 || m.ViewWidth != 2 {
		t.Fatal("resize")
	}
}
func TestInputFocusAndCallback(t *testing.T) {
	g := NewStyledRuneGrid(4, 4)
	m := NewModel(g, WithViewWidth(2), WithViewHeight(2))
	right := tea.KeyPressMsg{Code: tea.KeyRight}
	m, _ = m.Update(right)
	if m.Cursor() != (Point{}) {
		t.Fatal("blur")
	}
	m.Focus()
	m, _ = m.Update(right)
	m, _ = m.Update(tea.MouseWheelMsg{Button: tea.MouseWheelDown})
	if m.Cursor() != (Point{X: 1, Y: 1}) {
		t.Fatal(m.Cursor())
	}
	m.UpdateHandler = func(_ *Model, _ tea.Msg) { g.SetRune(Point{}, '!') }
	m, _ = m.Update(nil)
	if g.Cell(Point{}).Rune != '!' {
		t.Fatal("callback")
	}
}
func TestOversizedViewportOutput(t *testing.T) {
	m := NewModel(NewStyledRuneGrid(2, 2), WithViewHeight(4))
	if m.View() != "  \n  " {
		t.Fatalf("no trailing newline: %q", m.View())
	}
}

func TestStylesSurviveResizeAndShift(t *testing.T) {
	s := lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Background(lipgloss.Color("4"))
	m := NewStyledRuneGrid(3, 2, WithStyle(s))
	m.SetText(Point{}, "abc")
	m.Resize(2, 3)
	if m.Cell(Point{X: 1}).Style.Render("b") != s.Render("b") {
		t.Fatal("style lost on resize")
	}
	m.ShiftDown()
	m.ShiftRight()
	if c := m.Cell(Point{X: 1, Y: 1}); c.Rune != 'a' || c.Style.Render("a") != s.Render("a") {
		t.Fatal("shift content/style")
	}
	m.ShiftLeft()
	m.ShiftUp()
	if m.Cell(Point{}).Rune != 'a' {
		t.Fatal("shift back")
	}
	m.SetStyle(s.Bold(true))
	if !m.Cell(Point{}).Style.GetBold() {
		t.Fatal("bulk style")
	}
	m.Clear()
	if m.Cell(Point{}).Rune != 0 {
		t.Fatal("clear")
	}
}

func TestCellStyleCopyAndOptionOrdering(t *testing.T) {
	s := lipgloss.NewStyle().Bold(true)
	m := NewStyledRuneGrid(2, 1, WithStyle(s), WithContent([]string{"ab"}))
	copyStyle := m.GetCellStyle(Point{})
	if copyStyle == nil || !copyStyle.GetBold() {
		t.Fatal("content did not use preceding style option")
	}
	*copyStyle = copyStyle.Underline(true)
	if m.Cell(Point{}).Style.GetUnderline() {
		t.Fatal("GetCellStyle must return a copy")
	}
	m = NewStyledRuneGrid(2, 1, WithContent([]string{"ab"}), WithStyle(s))
	if m.Cell(Point{}).Style.GetBold() {
		t.Fatal("later default style must not restyle existing cells")
	}
	m.SetRune(Point{X: 1}, 'c')
	if !m.Cell(Point{X: 1}).Style.GetBold() {
		t.Fatal("new writes must use updated default")
	}
}
