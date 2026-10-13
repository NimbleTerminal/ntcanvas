package ntcanvas

import (
	"image"
	"strings"
	"testing"
	"unsafe"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

var (
	_ Surface = (*StyledRuneGrid)(nil)
	_ Surface = (*RuneGrid)(nil)
	_ Surface = (*GraphemeGrid)(nil)
)

func TestSurfaceContract(t *testing.T) {
	styled := NewStyledRuneGrid(4, 2)
	styled.SetText(Point{}, "abcd")
	runes := NewRuneGrid(4, 2)
	runes.SetText(Point{}, "abcd")
	graphemes := NewGraphemeGrid(4, 2)
	graphemes.SetText(Point{}, "abcd")
	for name, s := range map[string]Surface{"styled": styled, "rune": runes, "grapheme": graphemes} {
		t.Run(name, func(t *testing.T) {
			if got := s.Render(image.Rect(1, -2, 20, 1)); got != "bcd" {
				t.Fatal(got)
			}
			for _, r := range []image.Rectangle{{}, image.Rect(9, 9, 10, 10), {Min: Point{X: 3, Y: 2}, Max: Point{X: 1, Y: 1}}} {
				if s.Render(r) != "" {
					t.Fatal("empty render", r)
				}
			}
			m := NewModel(s, WithFocus(), WithViewWidth(2), WithViewHeight(1))
			copy := m
			copy, _ = copy.Update(tea.KeyPressMsg{Code: tea.KeyRight})
			if m.View() != "ab" || copy.View() != "bc" {
				t.Fatal("viewports share state", m.View(), copy.View())
			}
			if copy.Surface() != s {
				t.Fatal("surface changed")
			}
		})
	}
	// Numeric layout assertions avoid relying on architecture-specific byte counts.
	if unsafe.Sizeof(RuneCell{}) >= unsafe.Sizeof(Cell{})/4 {
		t.Fatal("rune cell is not compact")
	}
	if unsafe.Sizeof(GraphemeCell{}) >= unsafe.Sizeof(Cell{})/4 {
		t.Fatal("grapheme cell is not compact")
	}
}

type offsetSurface struct{ last image.Rectangle }

func (s *offsetSurface) Bounds() image.Rectangle         { return image.Rect(-5, 10, 5, 15) }
func (s *offsetSurface) Render(r image.Rectangle) string { s.last = r; return "custom" }
func TestModelCustomSurfaceAndResize(t *testing.T) {
	s := &offsetSurface{}
	m := NewModel(s, WithViewWidth(2), WithViewHeight(2))
	if m.View() != "custom" || s.last != image.Rect(-5, 10, -3, 12) {
		t.Fatal(s.last)
	}
	m.MoveRight(999)
	m.MoveDown(999)
	if m.Cursor() != (Point{X: 3, Y: 13}) {
		t.Fatal(m.Cursor())
	}
	m.MoveLeft(999)
	m.MoveUp(999)
	if m.Cursor() != s.Bounds().Min {
		t.Fatal(m.Cursor())
	}
	var zero Model
	zero.Focus()
	zero, _ = zero.Update(nil)
	if zero.View() != "" || !zero.Viewport().Empty() {
		t.Fatal("zero model")
	}
	g := NewRuneGrid(8, 4)
	m = NewModel(g, WithFocus(), WithViewWidth(3), WithViewHeight(2))
	m.SetCursor(Point{X: 7, Y: 3})
	other := m
	other.Blur()
	g.Resize(2, 1)
	g.SetText(Point{}, "ok")
	if m.Cursor() != (Point{X: 1}) || m.View() != "k" || !m.Focused() || other.Focused() {
		t.Fatal("resize/model state")
	}
	g.Resize(0, 0)
	m.MoveUp(9)
	m.MoveDown(9)
	if m.View() != "" {
		t.Fatal("empty")
	}
}

func TestCompactRuneStylesAndCopies(t *testing.T) {
	g := NewRuneGrid(5, 2)
	g.Style = CellStyle{Foreground: RGBColor(1, 2, 3), Attributes: Bold | Underline}
	g.SetText(Point{}, "abc")
	out := g.Render(g.Bounds())
	if strings.Count(out, "\x1b[1;4;38;2;1;2;3m") != 1 || ansi.Strip(out) != "abc  \n     " {
		t.Fatal(out)
	}
	g.SetTextWithStyle(Point{X: 1}, "Z", CellStyle{Background: IndexedColor(4)})
	g.CopyRect(Point{X: 1}, g, image.Rect(0, 0, 4, 1))
	if ansi.Strip(g.Render(g.Bounds())) != "aaZc \n     " {
		t.Fatal(g.Render(g.Bounds()))
	}
	g.SetStyleRect(image.Rect(0, 0, 2, 1), CellStyle{})
	if g.Cell(Point{}).Rune != 'a' || g.Cell(Point{}).Style != (CellStyle{}) || g.Style.Attributes != Bold|Underline {
		t.Fatal("style")
	}
	g.FillRect(image.Rect(0, 1, 5, 2), RuneCell{Rune: '\x1b'})
	if strings.Contains(ansi.Strip(g.Render(g.Bounds())), "\x1b") {
		t.Fatal("control escaped")
	}
	g.ClearRect(image.Rect(0, 1, 5, 2))
	g.Resize(2, 1)
	if ansi.Strip(g.Render(g.Bounds())) != "aa" {
		t.Fatal("resize")
	}
	g.Clear()
	if g.Cell(Point{}) != (RuneCell{}) {
		t.Fatal("clear")
	}
	if g.SetText(Point{Y: -1}, "x") || g.SetCell(Point{X: 9}, RuneCell{}) {
		t.Fatal("bounds")
	}
	minInt := -int(^uint(0)>>1) - 1
	if !g.SetText(Point{X: minInt}, "abc") {
		t.Fatal("clipped row")
	}
}

func TestCheckedGridConstructors(t *testing.T) {
	for name, newGrid := range map[string]func(int, int, SizeLimits) (Surface, error){
		"styled":   func(w, h int, l SizeLimits) (Surface, error) { return NewStyledRuneGridChecked(w, h, l) },
		"rune":     func(w, h int, l SizeLimits) (Surface, error) { return NewRuneGridChecked(w, h, l) },
		"grapheme": func(w, h int, l SizeLimits) (Surface, error) { return NewGraphemeGridChecked(w, h, l) },
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := newGrid(-1, 2, SizeLimits{}); err != ErrInvalidSize {
				t.Fatal(err)
			}
			if _, err := newGrid(3, 2, SizeLimits{MaxCells: 5}); err != ErrSizeLimit {
				t.Fatal(err)
			}
			s, err := newGrid(3, 2, SizeLimits{MaxCells: 6})
			if err != nil {
				t.Fatal(err)
			}
			g := s.(interface {
				ResizeChecked(int, int, SizeLimits) error
			})
			if err := g.ResizeChecked(4, 2, SizeLimits{MaxCells: 6}); err != ErrSizeLimit || s.Bounds() != image.Rect(0, 0, 3, 2) {
				t.Fatal("transaction", err)
			}
			if err := g.ResizeChecked(1, 1, SizeLimits{}); err != nil || s.Bounds() != image.Rect(0, 0, 1, 1) {
				t.Fatal(err)
			}
		})
	}
}

func TestEmptyOffsetViewportAndEdgeCursor(t *testing.T) {
	s := &offsetSurface{}
	m := NewModel(s)
	m.SetViewport(image.Rect(100, 100, 101, 101))
	if m.Cursor() != s.Bounds().Min || m.cursor != s.Bounds().Min || m.ViewWidth != 0 || m.ViewHeight != 0 || m.View() != "" {
		t.Fatal("empty viewport must anchor at bounds minimum", m.Cursor(), m.Viewport())
	}
	m.ViewWidth, m.ViewHeight = 3, 2
	m.SetCursor(Point{X: 100, Y: 100})
	if m.Cursor() != (Point{X: 4, Y: 14}) || m.Viewport() != image.Rect(4, 14, 5, 15) {
		t.Fatal("explicit partial edge viewport", m.Viewport())
	}
	m.MoveRight(1)
	m.MoveDown(1)
	if m.Cursor() != (Point{X: 2, Y: 13}) || m.Viewport() != image.Rect(2, 13, 5, 15) {
		t.Fatal("full viewport panning", m.Viewport())
	}
}

func TestRuneWithStyle(t *testing.T) {
	g := NewRuneGrid(2, 1)
	s := CellStyle{Attributes: Bold}
	if !g.SetRuneWithStyle(Point{}, '\x1b', s) || g.Cell(Point{}) != (RuneCell{Rune: '�', Style: s}) {
		t.Fatal("styled rune sanitization")
	}
	if g.SetRuneWithStyle(Point{X: 2}, 'x', s) {
		t.Fatal("out of bounds")
	}
	g.SetRuneWithStyle(Point{}, 0, s)
	if g.Cell(Point{}) != (RuneCell{Style: s}) {
		t.Fatal("zero rune")
	}
}

// Read methods must remain in the value method set, including for temporaries
// and map elements. Model is a child viewport, not a tea.Model root adapter.
var _ interface {
	View() string
	Bounds() image.Rectangle
	Init() tea.Cmd
	Update(tea.Msg) (Model, tea.Cmd)
} = Model{}

func TestModelValueReadsAndSize(t *testing.T) {
	g := NewRuneGrid(2, 1)
	g.SetText(Point{}, "ok")
	if NewModel(g).View() != "ok" || NewModel(g).Init() != nil {
		t.Fatal("temporary model reads")
	}
	models := map[string]Model{"grid": NewModel(g)}
	if models["grid"].Bounds() != g.Bounds() || models["grid"].View() != "ok" {
		t.Fatal("map value reads")
	}
	if unsafe.Sizeof(Model{}) > 128 {
		t.Fatalf("model grew to %d bytes", unsafe.Sizeof(Model{}))
	}
	t.Logf("Model: %d bytes; KeyMap: %d bytes", unsafe.Sizeof(Model{}), unsafe.Sizeof(KeyMap{}))
}
