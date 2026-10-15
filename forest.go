package main

import (
	tea "charm.land/bubbletea/v2"
	"fmt"
	"github.com/charmbracelet/x/ansi"
	"github.com/yaqyn/clibble-pokemon-cli/internal/brand"
	"strings"
	"time"
)

// A cell contains two vertical pixels. Foreground UI replaces the entire cell,
// so rain and scenery can never obscure text or the selected action.
type forestCell struct {
	top, bottom            int
	glyph                  rune
	foreground, background int
}
type forestScene struct {
	width, height int
	base          []forestCell
}

var forestPalette = func() []string {
	p := brand.Get().Forest
	result := []string{}
	for _, hex := range []string{p.Pine, p.Canopy, p.Fern, p.Leaf, p.Selection, p.Ivory} {
		var r, g, b int
		fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
		result = append(result, fmt.Sprintf("%d;%d;%d", r, g, b))
	}
	return result
}()

func (s *forestScene) prepare(width, height int) {
	if s.width == width && s.height == height && s.base != nil {
		return
	}
	s.width, s.height = width, height
	s.base = make([]forestCell, width*height)
	pixel := func(x, y, color int) {
		if x < 0 || x >= width || y < 0 || y >= height*2 {
			return
		}
		c := &s.base[(y/2)*width+x]
		if y%2 == 0 {
			c.top = color
		} else {
			c.bottom = color
		}
	}
	// Layered pine silhouettes frame a clearing; integer geometry stays crisp at
	// every terminal size. The scene is generated once, then reused until resize.
	for layer := 1; layer <= 3; layer++ {
		spacing := max(7, width/(8-layer))
		size := height + layer*3
		for x := -spacing / 2; x < width+spacing; x += spacing {
			if layer > 1 && x > width/4 && x < width*3/4 {
				continue
			}
			ground := height*2 - 3 + (x * 7 % 5)
			tip := ground - size - (x * x % 9)
			for y := tip; y < ground; y++ {
				half := (y - tip) / 4
				if (y-tip)%9 < 2 {
					half = max(0, half-2)
				}
				for dx := -half; dx <= half; dx++ {
					pixel(x+dx, y, layer)
				}
			}
			for y := ground - 3; y < height*2; y++ {
				pixel(x, y, 1)
			}
		}
	}
	for y := height*2 - 5; y < height*2; y++ {
		for x := 0; x < width; x++ {
			if (x*13+y*7)%9 < 4 {
				pixel(x, y, 2)
			}
		}
	}
}
func (s *forestScene) cells(frame int, motion bool) []forestCell {
	cells := append([]forestCell(nil), s.base...)
	if !motion {
		return cells
	}
	// Sparse drizzle, drifting motes, and a slow canopy shimmer. Arithmetic is
	// deterministic and independent from catch RNG, clocks, persistence, or I/O.
	for i := 0; i < max(5, s.width/4); i++ {
		x := (i*37 + frame/5) % s.width
		y := (i*11 + frame/2) % s.height
		c := &cells[y*s.width+x]
		c.glyph = '╱'
		c.foreground = 3
		c.background = c.bottom
	}
	for i := 0; i < 4; i++ {
		x := (i*29 + frame/6) % s.width
		y := max(0, s.height-4-(i*7+frame/12)%max(1, s.height-4))
		c := &cells[y*s.width+x]
		c.glyph = '·'
		c.foreground = 4
		c.background = c.bottom
	}
	for i := 0; i < len(cells); i++ {
		if cells[i].top == 3 && (i+frame/12)%31 == 0 {
			cells[i].top = 2
		}
	}
	return cells
}
func putForest(cells []forestCell, width, height, x, y int, text string, fg, bg int) {
	if y < 0 || y >= height {
		return
	}
	for _, r := range text {
		if x >= 0 && x < width {
			cells[y*width+x] = forestCell{glyph: r, foreground: fg, background: bg}
		}
		x++
	}
}
func forestRect(cells []forestCell, w, h, x, y, rw, rh int) {
	for row := y; row < y+rh; row++ {
		putForest(cells, w, h, x, row, strings.Repeat(" ", max(0, rw)), 5, 0)
	}
}
func renderForest(cells []forestCell, width, height int, color bool) string {
	var out strings.Builder
	lastFG, lastBG := -1, -1
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := cells[y*width+x]
			r, fg, bg := c.glyph, c.foreground, c.background
			if r == 0 {
				r = '▀'
				fg, bg = c.top, c.bottom
				if !color {
					r = ' '
					if fg > 0 || bg > 0 {
						r = '░'
					}
				}
			}
			if color && (lastFG != fg || lastBG != bg) {
				fmt.Fprintf(&out, "\x1b[38;2;%sm\x1b[48;2;%sm", forestPalette[fg], forestPalette[bg])
				lastFG, lastBG = fg, bg
			}
			out.WriteRune(r)
		}
		if color {
			out.WriteString("\x1b[0m")
			lastFG, lastBG = -1, -1
		}
		if y < height-1 {
			out.WriteByte('\n')
		}
	}
	return out.String()
}

type forestTick struct{}

func forestTimer() tea.Cmd {
	return tea.Tick(time.Second/12, func(time.Time) tea.Msg { return forestTick{} })
}
func (m *gameModel) animated() bool {
	return !m.cfg.noMotion && !m.cfg.progress.MotionOff && m.focused && m.cfg.view != nil && m.cfg.view.color
}
func (m *gameModel) animate() tea.Cmd {
	if m.animated() && !m.ticking {
		m.ticking = true
		return forestTimer()
	}
	return nil
}
func (m *gameModel) forestView() tea.View {
	if m.width < 40 || m.height < 18 {
		lines := strings.Split(ansi.Hardwrap("Clibble >_\nResize to 40 × 18.\nq: leave", max(1, m.width), false), "\n")
		v := tea.NewView(strings.Join(lines[:min(len(lines), max(1, m.height))], "\n"))
		v.AltScreen = true
		return v
	}
	w, h := min(m.width, 320), min(m.height, 120)
	m.scene.prepare(w, h)
	color := m.cfg.view != nil && m.cfg.view.color
	cells := m.scene.cells(m.frame, m.animated())
	text := func(x, y int, s string, fg, bg int) {
		putForest(cells, w, h, x, y, ansi.Truncate(s, max(0, w-x), "…"), fg, bg)
	}
	centered := func(y int, s string) { s = ansi.Truncate(s, w-6, "…"); text((w-ansi.StringWidth(s))/2, y, s, 5, 0) }
	p := m.cfg.progress.refreshed(m.cfg.now())
	text(2, 0, " Clibble >_ ", 4, 0)
	hud := fmt.Sprintf(" %d seeds  /  %d befriended ", p.Seeds, len(m.cfg.pokedex))
	text(max(2, w-len(hud)-2), 1, hud, 5, 0)
	panelWidth := min(54, w-6)
	left := (w - panelWidth) / 2
	rows := m.bodyRows()
	panelTop := 4
	forestRect(cells, w, h, left-1, panelTop-1, panelWidth+2, rows+3)
	centered(panelTop-1, m.heading())
	if m.screen == "record" || m.screen == "help" {
		lines := m.detailLines()
		offset := min(m.detailOffset, max(0, len(lines)-rows))
		for i := 0; i < rows && offset+i < len(lines); i++ {
			text(left+1, panelTop+1+i, ansi.Truncate(lines[offset+i], panelWidth-2, "…"), 5, 0)
		}
	} else {
		choices := m.choices()
		start := max(0, m.cursor-rows+1)
		visible := min(rows, len(choices)-start)
		top := panelTop + 1 + (rows-visible)/2
		for i := 0; i < visible; i++ {
			index := i + start
			label := "  " + choices[index].label
			fg, bg := 5, 0
			if index == m.cursor {
				label = "> " + choices[index].label
				fg, bg = 0, 4
			}
			label = ansi.Truncate(label, panelWidth-2, "…")
			label += " " + strings.Repeat(" ", max(0, panelWidth-3-ansi.StringWidth(label)))
			text(left+1, top+i, label, fg, bg)
		}
	}
	status := "You carry catch seeds. Approach gently."
	if m.screen == "encounter" {
		tries := 3 - p.Encounters[m.selected].Offers
		status = fmt.Sprintf("%d / 3 tries left  |  %d catch seeds", tries, p.Seeds)
		if e := p.Encounters[m.selected]; tries == 0 {
			status = "Let them rest: " + e.Until.Sub(m.cfg.now()).Round(time.Second).String()
		}
	}
	centered(h-8, status)
	// Narration always has its own opaque, fixed-height area, including at 40×18.
	forestRect(cells, w, h, 2, h-7, w-4, 7)
	text(2, h-7, "┌"+strings.Repeat("─", w-6)+"┐", 3, 0)
	lines := m.storyLines()
	page := m.storyPage % m.storyPages()
	for i := 0; i < 3; i++ {
		text(2, h-6+i, "│", 3, 0)
		text(w-3, h-6+i, "│", 3, 0)
		if i+page*3 < len(lines) {
			text(4, h-6+i, ansi.Truncate(lines[i+page*3], w-8, "…"), 5, 0)
		}
	}
	text(2, h-3, "└"+strings.Repeat("─", w-6)+"┘", 3, 0)
	footer := "↑↓: move  Enter: choose  Esc: back  q: quit"
	if w < 55 {
		footer = "↑↓  Enter: pick  Esc: back  q: quit"
	}
	if m.busy {
		footer = "You wait…  q: leave safely"
	} else if m.problem != "" {
		footer = "r: retry  Esc: back  q: quit"
	}
	centered(h-1, footer)
	if m.storyPages() > 1 {
		centered(h-2, fmt.Sprintf("Tab: read more (%d/%d)", page+1, m.storyPages()))
	} else if m.screen != "record" && m.screen != "help" {
		centered(h-2, fmt.Sprintf("%d / %d", m.cursor+1, len(m.choices())))
	}
	v := tea.NewView(renderForest(cells, w, h, color))
	v.AltScreen = true
	v.ReportFocus = true
	return v
}
