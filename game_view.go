package main

import (
	"bytes"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
)

func (m *gameModel) bodyRows() int { return max(4, min(m.height, 120)-13) }
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
	case "title":
		return "Clibble"
	case "settings":
		return "Your forest"
	case "travel":
		return "Your map"
	case "area":
		return placeName(m.location)
	case "encounter":
		return "You encounter " + placeName(m.selected)
	case "collection":
		return "Your field journal"
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
		return strings.Split("You choose your own pace.\n\nUp/down or j/k: move\nEnter or Space: choose\nEsc or Backspace: go back\nt: travel   c: collection\nLeft/right: map pages\nr: retry a failed action\nq or Ctrl+C: leave\n\nStronger wildling escape more often.\nEach offer uses one seed and one try.\nThree tries per wildling / 30 minutes.\nSeeds refill: one / 5 minutes, up to 12.\nForest settings can turn motion off.\nTravel position lasts this session.", "\n")
	}
	if m.screen == "collection" {
		if len(m.cfg.pokedex) == 0 {
			return strings.Split("Your first catch is still out there.\n\nChoose Travel from the trailhead,\nvisit a location, and approach\na wildling in the grass.", "\n")
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
			art = append(art, "", "You have caught this wildling.")
		} else {
			art = append(art, "", "You have not caught this wildling yet.")
		}
	}
	return art
}
func (m *gameModel) View() tea.View { return m.forestView() }
