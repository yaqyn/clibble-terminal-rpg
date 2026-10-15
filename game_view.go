package main

import (
	"bytes"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func (m *gameModel) bodyRows() int { return max(4, m.height-13) }
func (m *gameModel) storyLines() []string {
	story := m.narration
	if m.problem != "" {
		story = m.problem
	}
	return strings.Split(ansi.Wrap(story, max(1, m.width-8), ""), "\n")
}
func (m *gameModel) storyPages() int { return max(1, (len(m.storyLines())+2)/3) }
func (m *gameModel) heading() string {
	switch m.screen {
	case "travel":
		return "Your map"
	case "area":
		return placeName(m.location)
	case "encounter":
		return "You encounter " + placeName(m.selected)
	case "collection":
		return "Your collection"
	case "record":
		return placeName(m.record.Name) + " / Field record"
	case "help":
		return "Your field guide"
	default:
		return "The trailhead"
	}
}
func (m *gameModel) ink(s string) string {
	if m.cfg.view == nil {
		return s
	}
	return m.cfg.view.accent(s)
}
func (m *gameModel) menuLines(width, rows int) []string {
	choices := m.choices()
	lines := make([]string, 0, rows)
	start := 0
	if m.cursor >= rows {
		start = m.cursor - rows + 1
	}
	for i := start; i < len(choices) && len(lines) < rows; i++ {
		prefix := "  "
		if i == m.cursor {
			prefix = "> "
		}
		label := ansi.Truncate(prefix+choices[i].label, width, "…")
		if i == m.cursor {
			label = m.ink(label)
		}
		lines = append(lines, label)
	}
	return lines
}
func (m *gameModel) detailLines() []string {
	if m.screen == "help" {
		return m.sideLines(max(1, m.width-4))
	}
	return m.recordLines()
}
func (m *gameModel) recordLines() []string {
	var output bytes.Buffer
	view := terminalView{out: &output, width: 65}
	view.inspect(m.record)
	return strings.Split(strings.TrimSpace(output.String()), "\n")
}
func (m *gameModel) sideLines(width int) []string {
	if m.screen == "help" {
		return strings.Split("You choose your own pace.\n\nUp/down or j/k: move\nEnter or Space: choose\nEsc or Backspace: go back\nt: travel   c: collection\nLeft/right: map pages\nr: retry a failed action\nq or Ctrl+C: leave\n\nStronger Pokemon escape more often.\nYour successful catches save at once.\nTravel position lasts this session.", "\n")
	}
	if m.screen == "collection" {
		if len(m.cfg.pokedex) == 0 {
			return strings.Split("Your first catch is still out there.\n\nChoose Travel from the trailhead,\nvisit a location, and approach\na Pokemon in the grass.", "\n")
		}
		choices := m.choices()
		if m.cursor < len(choices) && choices[m.cursor].id != "back" {
			p := m.cfg.pokedex[choices[m.cursor].id]
			var out bytes.Buffer
			v := terminalView{out: &out, width: 40}
			v.inspect(p)
			return strings.Split(strings.TrimSpace(out.String()), "\n")
		}
	}
	art := []string{"       .       *", "        /\\       /\\", "   ,   /  \\     /  \\", "  /|\\    . . .     ,", "      . >_    .   /|\\", "  ,  .     /\\  .", " /|\\      /  \\    ,", "              . /|\\"}
	if m.busy {
		return append(art, "", "You wait for the trail to clear…")
	}
	choices := m.choices()
	if m.cursor < len(choices) {
		art = append(art, "", ansi.Wrap(choices[m.cursor].hint, width, ""))
	}
	if m.screen == "encounter" {
		if _, ok := m.cfg.pokedex[m.selected]; ok {
			art = append(art, "", "You have caught this Pokemon.")
		} else {
			art = append(art, "", "You have not caught this Pokemon yet.")
		}
	}
	return art
}
func (m *gameModel) View() tea.View {
	width := max(1, m.width-4)
	if m.width < 40 || m.height < 18 {
		content := ansi.Hardwrap("Clibble >_\nResize to 40 × 18.\nq: leave", max(1, m.width), false)
		lines := strings.Split(content, "\n")
		if len(lines) > max(1, m.height) {
			lines = lines[:max(1, m.height)]
		}
		v := tea.NewView(strings.Join(lines, "\n"))
		v.AltScreen = true
		return v
	}
	saveMode := "Auto-saved"
	if m.cfg.save == nil {
		saveMode = "Temporary session"
	}
	lines := []string{m.ink("Clibble >_") + "   A little wild in your terminal.", fmt.Sprintf("%d caught  /  %s", len(m.cfg.pokedex), saveMode), m.ink(m.heading()), strings.Repeat("─", width)}
	rows := m.bodyRows()
	if m.screen == "record" || m.screen == "help" {
		record := m.detailLines()
		mOffset := min(m.detailOffset, max(0, len(record)-rows))
		for i := 0; i < rows; i++ {
			line := ""
			if i+mOffset < len(record) {
				line = record[i+mOffset]
			}
			lines = append(lines, ansi.Truncate(line, width, "…"))
		}
	} else {
		menuWidth := width
		wide := width >= 84
		if wide {
			menuWidth = width/2 - 2
		}
		menu := m.menuLines(menuWidth, rows)
		side := m.sideLines(width - menuWidth - 4)

		for i := 0; i < rows; i++ {
			left := ""
			if i < len(menu) {
				left = menu[i]
			}
			if wide {
				right := ""
				if i < len(side) {
					right = ansi.Truncate(side[i], width-menuWidth-4, "…")
				}
				left += strings.Repeat(" ", max(0, menuWidth-ansi.StringWidth(left))) + "  │ " + right
			}
			lines = append(lines, left)
		}
	}
	wrapped := m.storyLines()
	page := m.storyPage % m.storyPages()
	lines = append(lines, "┌"+strings.Repeat("─", width-2)+"┐")
	for i := 0; i < 3; i++ {
		s := ""
		if i+page*3 < len(wrapped) {
			s = wrapped[i+page*3]
		}
		s = ansi.Truncate(s, width-4, "…")
		lines = append(lines, "│ "+s+strings.Repeat(" ", max(0, width-4-ansi.StringWidth(s)))+" │")
	}
	lines = append(lines, "└"+strings.Repeat("─", width-2)+"┘")
	if m.storyPages() > 1 {
		lines = append(lines, fmt.Sprintf("Tab: read more (%d/%d)", page+1, m.storyPages()))
	}
	footer := "↑↓ / j k: move   Enter: choose   Esc: back   q: leave"
	if m.screen == "travel" {
		footer = "↑↓: choose  Enter: travel  ←→: pages  Esc: back  q: leave"
	}
	if m.screen == "record" || m.screen == "help" {
		footer = "↑↓: scroll  Enter / Esc: back  q: leave"
	}
	if width < 70 {
		footer = "↑↓  Enter: pick  Esc: back  q: quit"
		if m.screen == "record" || m.screen == "help" {
			footer = "↑↓: scroll  Esc: back  q: quit"
		}
	}
	if m.busy {
		footer = "Loading…  Esc: cancel  q: quit"
		if m.pending == "catch" {
			footer = "Settling your catch… q: leave safely"
		}
	}
	if m.problem != "" {
		footer = "r: retry   Esc: back   q: leave"
	}
	if m.screen != "record" && m.screen != "help" && width >= 45 {
		footer = fmt.Sprintf("%d/%d  ", m.cursor+1, len(m.choices())) + footer
	}
	lines = append(lines, footer)
	for i, line := range lines {
		lines[i] = "  " + ansi.Truncate(line, width, "…")
	}
	v := tea.NewView(strings.Join(lines, "\n"))
	v.AltScreen = true
	// Preserve the user's terminal background and font. NO_COLOR controls accents.
	return v
}
