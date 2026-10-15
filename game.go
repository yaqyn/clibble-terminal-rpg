package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	tea "charm.land/bubbletea/v2"
	"github.com/yaqyn/clibble/internal/pokeapi"
	"golang.org/x/term"
)

// gameTasks starts work before returning its receive command, so shutdown can
// drain even tasks whose messages Bubble Tea no longer needs. No worker writes
// to the model; results are applied by Update alone.
type gameTasks struct {
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func newGameTasks(ctx context.Context) *gameTasks {
	ctx, cancel := context.WithCancel(ctx)
	return &gameTasks{ctx: ctx, cancel: cancel}
}
func (t *gameTasks) start(work func() tea.Msg) tea.Cmd {
	result := make(chan tea.Msg, 1)
	t.wg.Add(1)
	go func() { defer t.wg.Done(); result <- work() }()
	return func() tea.Msg { return <-result }
}
func (t *gameTasks) close() { t.cancel(); t.wg.Wait() }

type gameChoice struct{ id, label, hint string }
type gameResult struct {
	id        int
	kind      string
	locations pokeapi.Locations
	area      pokeapi.Area
	catch     catchResult
	err       error
}
type gameModel struct {
	cfg                                 *config
	tasks                               *gameTasks
	screen                              string
	cursor, width, height, detailOffset int
	areaCursor, recordCursor            int
	location, selected, recordFrom      string
	locations                           pokeapi.Locations
	encounters                          []string
	record                              pokeapi.Pokemon
	narration, problem                  string
	busy                                bool
	pending                             string
	requestID                           int
	cancel                              context.CancelFunc
	retry                               func() tea.Cmd
}

func newGame(cfg *config, tasks *gameTasks) *gameModel {
	return &gameModel{cfg: cfg, tasks: tasks, screen: "camp", width: 80, height: 24, narration: "You stand at the trailhead. Somewhere beyond the tall grass, your next catch is waiting. Where will you go?"}
}
func (m *gameModel) Init() tea.Cmd { return nil }
func (m *gameModel) request(kind, message string, work func(context.Context) gameResult) tea.Cmd {
	if m.cancel != nil {
		m.cancel()
	}
	ctx, cancel := context.WithCancel(m.tasks.ctx)
	m.cancel = cancel
	m.requestID++
	id := m.requestID
	m.busy = true
	m.pending = kind
	m.problem = ""
	m.narration = message
	return m.tasks.start(func() tea.Msg { result := work(ctx); result.id = id; result.kind = kind; return result })
}
func (m *gameModel) travel(page *string) tea.Cmd {
	m.retry = func() tea.Cmd { return m.travel(page) }
	return m.request("travel", "You unfold your map and look for a new trail…", func(ctx context.Context) gameResult {
		result, err := m.cfg.client.ListLocations(ctx, page)
		return gameResult{locations: result, err: err}
	})
}
func (m *gameModel) explore(name string) tea.Cmd {
	m.retry = func() tea.Cmd { return m.explore(name) }
	return m.request("explore", "You follow the trail toward "+placeName(name)+"…", func(ctx context.Context) gameResult {
		result, err := m.cfg.client.Explore(ctx, name)
		return gameResult{area: result, err: err}
	})
}
func (m *gameModel) throw() tea.Cmd {
	name := m.selected
	owned := copyCollection(m.cfg.pokedex)
	m.retry = m.throw
	return m.request("catch", "You throw a Pokeball toward "+placeName(name)+". You watch it closely…", func(ctx context.Context) gameResult {
		result, err := attemptCatch(ctx, m.cfg.client, owned, m.cfg.random, m.cfg.save, name)
		return gameResult{catch: result, err: err}
	})
}
func (m *gameModel) move(screen, narration string) {
	m.screen = screen
	m.cursor = 0
	m.detailOffset = 0
	m.problem = ""
	m.narration = narration
}
func (m *gameModel) camp() {
	m.move("camp", "You pause at the trailhead. Your next adventure starts with a choice.")
}
func (m *gameModel) back() {
	switch m.screen {
	case "record":
		m.move(m.recordFrom, "You close the field record. What will you do next?")
		m.cursor = m.recordCursor
	case "encounter":
		m.move("area", "You return to the grass. You look for another Pokemon.")
		m.cursor = m.areaCursor
	default:
		m.camp()
	}
}
func (m *gameModel) openCollection() {
	m.move("collection", fmt.Sprintf("You open your collection. You have caught %d distinct Pokemon.", len(m.cfg.pokedex)))
	if len(m.cfg.pokedex) == 0 {
		m.narration = "Your collection is waiting for its first story. Travel to a location and choose a Pokemon to catch."
	}
}
func (m *gameModel) choices() []gameChoice {
	switch m.screen {
	case "camp":
		leave := gameChoice{"quit", "Save and leave", "Your successful catches are already saved"}
		if m.cfg.save == nil {
			leave = gameChoice{"quit", "Leave this trail", "Your temporary collection ends with this session"}
		}
		return []gameChoice{{"travel", "Travel", "Choose a location to visit"}, {"collection", "Your collection", "Inspect the Pokemon you have caught"}, {"help", "Field guide", "Learn the controls and catch rules"}, leave}
	case "travel":
		choices := make([]gameChoice, 0, len(m.locations.Results)+3)
		for _, area := range m.locations.Results {
			choices = append(choices, gameChoice{area.Name, placeName(area.Name), "Enter to travel here and explore"})
		}
		if m.locations.Next != nil {
			choices = append(choices, gameChoice{"next", "More destinations", "Next page of the map"})
		}
		if m.locations.Previous != nil {
			choices = append(choices, gameChoice{"previous", "Previous destinations", "Previous page of the map"})
		}
		return append(choices, gameChoice{"back", "Return to trailhead", "Choose another activity"})
	case "area":
		choices := make([]gameChoice, 0, len(m.encounters)+3)
		for _, name := range m.encounters {
			label := placeName(name)
			if _, ok := m.cfg.pokedex[name]; ok {
				label += " [caught]"
			}
			choices = append(choices, gameChoice{name, label, "Enter to approach this Pokemon"})
		}
		return append(choices, gameChoice{"explore", "Explore again", "Look around this location again"}, gameChoice{"travel", "Travel onward", "Open your map"}, gameChoice{"back", "Return to trailhead", "Take a moment to plan"})
	case "encounter":
		if _, ok := m.cfg.pokedex[m.selected]; ok {
			return []gameChoice{{"inspect", "Inspect your catch", "Read its saved field record"}, {"back", "Return to the grass", "Look for another Pokemon"}, {"collection", "Your collection", "Browse your caught Pokemon"}}
		}
		return []gameChoice{{"throw", "Throw a Pokeball", "Attempt a catch; some Pokemon may escape"}, {"back", "Return to the grass", "Leave this encounter"}}
	case "collection":
		choices := make([]gameChoice, 0, len(m.cfg.pokedex)+1)
		for _, name := range sortedKeys(m.cfg.pokedex) {
			choices = append(choices, gameChoice{name, placeName(name), "Read this record offline"})
		}
		return append(choices, gameChoice{"back", "Return to trailhead", "Choose your next trail"})
	default:
		return []gameChoice{{"back", "Back", "Return to your previous screen"}}
	}
}
func (m *gameModel) choose() tea.Cmd {
	choices := m.choices()
	if m.cursor < 0 || m.cursor >= len(choices) {
		return nil
	}
	choice := choices[m.cursor]
	switch choice.id {
	case "quit":
		return tea.Quit
	case "back":
		m.back()
	case "travel":
		return m.travel(nil)
	case "next":
		return m.travel(m.locations.Next)
	case "previous":
		return m.travel(m.locations.Previous)
	case "explore":
		return m.explore(m.location)
	case "throw":
		return m.throw()
	case "help":
		m.move("help", "You open your field guide. You can take every trail at your own pace.")
	case "collection":
		m.openCollection()
	case "inspect":
		m.recordCursor = m.cursor
		m.record = m.cfg.pokedex[m.selected]
		m.recordFrom = "encounter"
		m.move("record", "You take a closer look at "+placeName(m.record.Name)+". This record is yours to read offline.")
	default:
		switch m.screen {
		case "travel":
			m.selected = choice.id
			return m.explore(choice.id)
		case "area":
			m.areaCursor = m.cursor
			m.selected = choice.id
			m.move("encounter", "You spot "+placeName(choice.id)+" in the grass. You move a little closer. What will you do?")
		case "collection":
			m.recordCursor = m.cursor
			m.record = m.cfg.pokedex[choice.id]
			m.recordFrom = "collection"
			m.move("record", "You open your saved record for "+placeName(choice.id)+".")
		}
	}
	return nil
}
func (m *gameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case gameResult:
		if msg.id != m.requestID {
			return m, nil
		}
		m.busy = false
		m.pending = ""
		if m.cancel != nil {
			m.cancel()
			m.cancel = nil
		}
		if msg.err != nil {
			m.problem = fmt.Sprintf("%v. Press r to retry, or Esc to go back.", msg.err)
			m.narration = "You pause on the trail. That action could not be completed; your collection has not changed."
			return m, nil
		}
		switch msg.kind {
		case "travel":
			m.locations = msg.locations
			m.move("travel", "You trace the paths on your map. Choose where you want to travel.")
		case "explore":
			// Location changes only after a successful request, so failed travel keeps
			// the old area and its encounter list together.
			if m.screen == "travel" {
				m.location = m.selected
			}
			m.encounters = nil
			seen := map[string]bool{}
			for _, e := range msg.area.PokemonEncounters {
				if !seen[e.Pokemon.Name] {
					m.encounters = append(m.encounters, e.Pokemon.Name)
					seen[e.Pokemon.Name] = true
				}
			}
			m.move("area", "You step into "+placeName(m.location)+". You notice movement in the grass. Choose a Pokemon to approach.")
			if len(m.encounters) == 0 {
				m.narration = "You look around " + placeName(m.location) + ". You find no Pokemon here. You can explore again or travel onward."
			}
		case "catch":
			p := msg.catch.pokemon
			if msg.catch.caught {
				m.cfg.pokedex = msg.catch.owned
				m.narration = "You caught " + placeName(p.Name) + "! You add its record to your collection."
				if m.cfg.save != nil {
					m.narration += " Your catch is saved."
				} else {
					m.narration += " This is a temporary session."
				}
				m.cursor = 0
			} else if msg.catch.already {
				m.narration = "You already have " + placeName(p.Name) + " in your collection. You can inspect its record."
			} else {
				m.narration = "You watch " + placeName(p.Name) + " slip free! You still have another chance. Will you try again?"
			}
		}
	case tea.KeyPressMsg:
		key := msg.String()
		if key == "q" || key == "ctrl+c" {
			return m, tea.Quit
		}
		if m.width < 40 || m.height < 18 {
			return m, nil
		}
		if m.busy {
			if key == "esc" && m.pending != "catch" {
				if m.cancel != nil {
					m.cancel()
				}
				m.requestID++
				m.busy = false
				m.pending = ""
				m.narration = "You put your map away. You can choose another path."
				m.problem = ""
			}
			return m, nil
		}
		switch key {
		case "esc", "backspace":
			m.back()
		case "r":
			if m.problem != "" && m.retry != nil {
				return m, m.retry()
			}
		case "up", "k":
			if m.screen == "record" || m.screen == "help" {
				m.detailOffset = max(0, m.detailOffset-1)
			} else {
				m.cursor = max(0, m.cursor-1)
			}
		case "down", "j":
			if m.screen == "record" || m.screen == "help" {
				m.detailOffset = min(max(0, len(m.detailLines())-m.bodyRows()), m.detailOffset+1)
			} else {
				m.cursor = min(len(m.choices())-1, m.cursor+1)
			}
		case "home":
			m.cursor = 0
			m.detailOffset = 0
		case "end":
			m.cursor = len(m.choices()) - 1
			if m.screen == "record" || m.screen == "help" {
				m.detailOffset = max(0, len(m.detailLines())-m.bodyRows())
			}
		case "pgdown":
			m.cursor = min(len(m.choices())-1, m.cursor+m.bodyRows())
		case "pgup":
			m.cursor = max(0, m.cursor-m.bodyRows())
		case "left", "p":
			if m.screen == "travel" && m.locations.Previous != nil {
				return m, m.travel(m.locations.Previous)
			}
		case "right", "n":
			if m.screen == "travel" && m.locations.Next != nil {
				return m, m.travel(m.locations.Next)
			}
		case "enter", " ":
			return m, m.choose()
		case "t":
			return m, m.travel(nil)
		case "c":
			m.openCollection()
		case "?":
			m.move("help", "You open your field guide.")
		}
	}
	return m, nil
}
func supportsGame(in io.Reader, out io.Writer) bool {
	input, a := in.(*os.File)
	output, b := out.(*os.File)
	return a && b && term.IsTerminal(int(input.Fd())) && term.IsTerminal(int(output.Fd())) && os.Getenv("TERM") != "dumb"
}
func runGame(cfg *config, in io.Reader, out io.Writer) error {
	tasks := newGameTasks(cfg.ctx)
	defer tasks.close()
	_, err := tea.NewProgram(newGame(cfg, tasks), tea.WithContext(cfg.ctx), tea.WithInput(in), tea.WithOutput(out)).Run()
	if cfg.ctx.Err() != nil {
		return cfg.ctx.Err()
	}
	if errors.Is(err, tea.ErrProgramKilled) {
		return context.Canceled
	}
	return err
}
func placeName(name string) string {
	words := strings.Fields(strings.ReplaceAll(name, "-", " "))
	for i, w := range words {
		words[i] = strings.ToUpper(w[:1]) + w[1:]
	}
	return strings.Join(words, " ")
}
