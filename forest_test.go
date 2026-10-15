package main

import (
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"strings"
	"testing"
	"time"
)

func TestForestMenuIsCenteredAndSceneryCannotCoverSelection(t *testing.T) {
	m, _ := gameForTest(t)
	m.width, m.height = 80, 24
	m.cfg.view = &terminalView{color: true}
	content := m.View().Content
	if !strings.Contains(content, "\x1b[48;2;177;209;142m") {
		t.Fatal("selected leaf fill missing")
	}
	for _, line := range strings.Split(ansi.Strip(content), "\n") {
		if strings.Contains(line, "> Travel") && strings.Index(line, "> Travel") < 15 {
			t.Fatal("menu is not centered")
		}
	}
	before := m.cfg.progress
	m.Update(forestTick{})
	if m.cfg.progress.Balls != before.Balls || len(m.cfg.progress.Encounters) != len(before.Encounters) {
		t.Fatal("weather changed gameplay")
	}
	if !strings.Contains(m.View().Content, "> Travel") {
		t.Fatal("rain hid selection")
	}
}
func TestMotionSettingPersistsAndFocusStopsTicks(t *testing.T) {
	m, _ := gameForTest(t)
	m.cfg.view = &terminalView{color: true}
	m.cfg.progress = freshProgress(time.Now())
	var saved gameProgress
	m.cfg.saveProgress = func(_ mapPokemon, p gameProgress) error { saved = p; return nil }
	m.move("settings", "Choose motion")
	gameComplete(t, m, gameKey(m, "enter"))
	if !saved.MotionOff || m.animated() {
		t.Fatal("motion preference not applied")
	}
	m.cfg.progress.MotionOff = false
	m.focused = true
	if m.Init() == nil {
		t.Fatal("no animation scheduled")
	}
	m.Update(tea.BlurMsg{})
	_, cmd := m.Update(forestTick{})
	if cmd != nil {
		t.Fatal("background terminal keeps animating")
	}
	_, cmd = m.Update(tea.FocusMsg{})
	if cmd == nil {
		t.Fatal("focus did not resume animation")
	}
	m.cfg.noMotion = true
	if m.animated() {
		t.Fatal("flag ignored")
	}
}
func TestTitleStartsWithoutNetworkOrSpending(t *testing.T) {
	cfg, _, api := testConfig()
	tasks := newGameTasks(cfg.ctx)
	defer tasks.close()
	m := newGame(cfg, tasks)
	if m.screen != "title" || m.choices()[0].label != "Start exploring" {
		t.Fatal("title missing")
	}
	gameKey(m, "enter")
	if m.screen != "camp" || api.calls != 0 {
		t.Fatal("starting made requests")
	}
}
func BenchmarkForestFrame(b *testing.B) {
	scene := forestScene{}
	scene.prepare(100, 30)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderForest(scene.cells(i, true), 100, 30, true)
	}
}
