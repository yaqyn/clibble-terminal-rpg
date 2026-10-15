package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/yaqyn/clibble-pokemon-cli/internal/pokeapi"
)

type trailAPI struct {
	fakeAPI
	area pokeapi.Area
}

func (a *trailAPI) Explore(context.Context, string) (pokeapi.Area, error) { return a.area, a.err }
func gameForTest(t *testing.T) (*gameModel, *trailAPI) {
	t.Helper()
	cfg, _, f := testConfig()
	api := &trailAPI{fakeAPI: *f}
	if err := json.Unmarshal([]byte(`{"pokemon_encounters":[{"pokemon":{"name":"pidgey"}}]}`), &api.area); err != nil {
		t.Fatal(err)
	}
	api.page = pokeapi.Locations{Results: []pokeapi.Resource{{Name: "kanto-route-1-area"}}}
	cfg.client = api
	tasks := newGameTasks(cfg.ctx)
	t.Cleanup(tasks.close)
	return newGame(cfg, tasks), api
}
func gameKey(m *gameModel, key string) tea.Cmd {
	code := rune(0)
	switch key {
	case "enter":
		code = tea.KeyEnter
	case "esc":
		code = tea.KeyEscape
	case "down":
		code = tea.KeyDown
	case "up":
		code = tea.KeyUp
	case "tab":
		code = tea.KeyTab
	default:
		code = []rune(key)[0]
	}
	_, cmd := m.Update(tea.KeyPressMsg{Code: code})
	return cmd
}
func gameComplete(t *testing.T, m *gameModel, cmd tea.Cmd) {
	t.Helper()
	if cmd == nil {
		t.Fatal("expected background work")
	}
	m.Update(cmd())
}
func TestGameTravelCatchInspectAndReturn(t *testing.T) {
	m, _ := gameForTest(t)
	gameComplete(t, m, gameKey(m, "enter"))
	if m.screen != "travel" {
		t.Fatal(m.screen)
	}
	gameComplete(t, m, gameKey(m, "enter"))
	if m.screen != "area" || m.location != "kanto-route-1-area" {
		t.Fatal(m.screen, m.location)
	}
	gameKey(m, "enter")
	if m.screen != "encounter" || m.selected != "pidgey" {
		t.Fatal(m.screen)
	}
	gameComplete(t, m, gameKey(m, "enter"))
	if len(m.cfg.pokedex) != 1 || !strings.Contains(m.narration, "You caught Pidgey!") {
		t.Fatal(m.narration)
	}
	gameKey(m, "enter")
	if m.screen != "record" || m.record.Name != "pidgey" {
		t.Fatal(m.screen)
	}
	gameKey(m, "esc")
	if m.screen != "encounter" {
		t.Fatal(m.screen)
	}
	gameKey(m, "esc")
	if m.screen != "area" {
		t.Fatal(m.screen)
	}
}
func TestGameCatchFailureEscapeAndRetry(t *testing.T) {
	m, _ := gameForTest(t)
	m.screen = "encounter"
	m.selected = "pidgey"
	m.cfg.save = func(map[string]pokeapi.Pokemon) error { return errors.New("disk full") }
	gameComplete(t, m, gameKey(m, "enter"))
	if len(m.cfg.pokedex) != 0 || !strings.Contains(m.problem, "disk full") {
		t.Fatal("save failure was hidden", m.problem)
	}
	m.cfg.save = nil
	m.cfg.random = func() float64 { return 0.99 }
	gameComplete(t, m, gameKey(m, "r"))
	if len(m.cfg.pokedex) != 0 || !strings.Contains(m.narration, "slip free") {
		t.Fatal("escape retained a catch", m.narration)
	}
	m.cfg.random = func() float64 { return 0 }
	gameComplete(t, m, gameKey(m, "enter"))
	if len(m.cfg.pokedex) != 1 {
		t.Fatal("retry did not catch")
	}
}
func TestGameCancellationIgnoresStaleResults(t *testing.T) {
	m, _ := gameForTest(t)
	cmd := gameKey(m, "enter")
	id := m.requestID
	gameKey(m, "esc")
	m.Update(cmd())
	if m.busy || m.screen != "camp" {
		t.Fatal("cancelled result changed screen")
	}
	m.Update(gameResult{id: id, kind: "catch", catch: catchResult{caught: true, owned: map[string]pokeapi.Pokemon{"pidgey": {Name: "pidgey"}}}})
	if len(m.cfg.pokedex) != 0 {
		t.Fatal("stale catch result changed collection")
	}
}
func TestGameTravelFailurePreservesCurrentLocation(t *testing.T) {
	m, api := gameForTest(t)
	m.location = "old-area"
	m.screen = "travel"
	m.selected = "new-area"
	api.err = errors.New("offline")
	gameComplete(t, m, m.explore(m.selected))
	if m.location != "old-area" || m.problem == "" {
		t.Fatal("failed travel changed location")
	}
	api.err = nil
	gameComplete(t, m, gameKey(m, "r"))
	if m.location != "new-area" {
		t.Fatal("retry failed")
	}
}
func TestGameTaskShutdownWaitsForSaving(t *testing.T) {
	m, _ := gameForTest(t)
	m.screen = "encounter"
	m.selected = "pidgey"
	started := make(chan struct{})
	finish := make(chan struct{})
	saved := make(chan struct{})
	m.cfg.save = func(map[string]pokeapi.Pokemon) error { close(started); <-finish; close(saved); return nil }
	cmd := gameKey(m, "enter")
	<-started
	if gameKey(m, "enter") != nil {
		t.Fatal("double catch allowed")
	}
	if gameKey(m, "q") == nil {
		t.Fatal("quit unavailable while saving")
	}
	closed := make(chan struct{})
	go func() { m.tasks.close(); close(closed) }()
	select {
	case <-closed:
		t.Fatal("shutdown did not wait for save")
	case <-time.After(10 * time.Millisecond):
	}
	close(finish)
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("shutdown hung")
	}
	<-saved
	m.Update(cmd())
}
func TestGameViewFitsSmallAndWideScreens(t *testing.T) {
	m, _ := gameForTest(t)
	for _, size := range [][2]int{{20, 8}, {40, 18}, {60, 24}, {110, 32}} {
		m.width = size[0]
		m.height = size[1]
		for _, screen := range []string{"camp", "travel", "area", "encounter", "collection", "record", "help"} {
			m.screen = screen
			m.record = pokeapi.Pokemon{Name: "pidgey"}
			m.selected = "pidgey"
			view := m.View()
			content := view.Content
			lines := strings.Split(content, "\n")
			if len(lines) > m.height {
				t.Errorf("%s %v: %d lines", screen, size, len(lines))
			}
			for _, line := range lines {
				if ansi.StringWidth(line) > m.width {
					t.Errorf("%s %v: line too wide: %q", screen, size, line)
				}
			}
		}
	}
}
func TestGameMenusScrollAndCollectionWorksOffline(t *testing.T) {
	m, api := gameForTest(t)
	m.height = 18
	m.width = 60
	m.screen = "collection"
	for i := 0; i < 100; i++ {
		name := fmt.Sprintf("pokemon-%d", i)
		m.cfg.pokedex[name] = pokeapi.Pokemon{Name: name}
	}
	for i := 0; i < 80; i++ {
		gameKey(m, "down")
	}
	selected := m.choices()[m.cursor].label
	if !strings.Contains(m.View().Content, selected) {
		t.Fatal("selection scrolled off screen")
	}
	gameKey(m, "enter")
	if m.screen != "record" || api.calls != 0 {
		t.Fatal("inspection made a network request")
	}
	gameKey(m, "esc")
	if m.cursor != 80 || m.screen != "collection" {
		t.Fatal("return did not restore the selected record", m.cursor)
	}
}

func TestGameSmallScreenDoesNotActivateHiddenChoices(t *testing.T) {
	m, api := gameForTest(t)
	m.width = 20
	if gameKey(m, "enter") != nil || api.calls != 0 {
		t.Fatal("hidden choice was activated")
	}
	if gameKey(m, "q") == nil {
		t.Fatal("small screen prevented quitting")
	}
}

func TestGameNarrowNarrationCanBeReadInFull(t *testing.T) {
	m, _ := gameForTest(t)
	m.width, m.height = 40, 18
	if !strings.Contains(m.View().Content, "Tab: read more") {
		t.Fatal("long narration lacks a continuation hint")
	}
	gameKey(m, "tab")
	if !strings.Contains(m.View().Content, "Where will you go?") || m.screen != "camp" {
		t.Fatal("continuation lost the story or activated a choice")
	}
}
