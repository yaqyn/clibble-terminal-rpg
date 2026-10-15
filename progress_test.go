package main

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestCatchBudgetSurvivesRestartAndEscape(t *testing.T) {
	cfg, _, _ := testConfig()
	cfg.random = func() float64 { return 1 }
	now := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	path := filepath.Join(t.TempDir(), "game.json")
	cfg.progress = freshProgress(now)
	cfg.clock = func() time.Time { return now }
	cfg.saveProgress = func(p mapPokemon, g gameProgress) error { return saveGame(path, p, g) }
	for i := 0; i < 3; i++ {
		if err := commandCatch(cfg, "pidgey"); err != nil {
			t.Fatal(err)
		}
		p, g, err := loadGame(path)
		if err != nil {
			t.Fatal(err)
		}
		cfg.pokedex, cfg.progress = p, g
	}
	if cfg.progress.Balls != 9 {
		t.Fatal(cfg.progress)
	}
	if err := commandCatch(cfg, "pidgey"); err == nil {
		t.Fatal("fourth throw permitted")
	}
	now = now.Add(encounterWindow)
	if err := commandCatch(cfg, "pidgey"); err != nil {
		t.Fatal(err)
	}
	if cfg.progress.Balls != 11 || cfg.progress.Encounters["pidgey"].Throws != 1 {
		t.Fatal(cfg.progress)
	}
}
func TestCatchBudgetRollsBackOnFailure(t *testing.T) {
	cfg, _, _ := testConfig()
	now := time.Now()
	cfg.progress = freshProgress(now)
	cfg.saveProgress = func(mapPokemon, gameProgress) error { return errors.New("disk full") }
	_, err := attemptCatch(context.Background(), cfg.client, cfg.pokedex, func() float64 { return 1 }, cfg.persist, cfg.progress, now, "pidgey")
	if err == nil || cfg.progress.Balls != 12 || len(cfg.progress.Encounters) != 0 {
		t.Fatal("failed transaction changed budget")
	}
}
func TestLegacySaveUpgradesWithoutLosingCollection(t *testing.T) {
	cfg, _, _ := testConfig()
	path := filepath.Join(t.TempDir(), "game.json")
	if err := savePokedex(path, cfg.pokedex); err != nil {
		t.Fatal(err)
	}
	p, g, err := loadGame(path)
	if err != nil {
		t.Fatal(err)
	}
	if g.Balls != 12 {
		t.Fatal(g)
	}
	g.Balls = 4
	g.MotionOff = true
	if err := saveGame(path, p, g); err != nil {
		t.Fatal(err)
	}
	_, g, err = loadGame(path)
	if err != nil || g.Balls != 4 || !g.MotionOff {
		t.Fatal(g, err)
	}
	if _, err := loadPokedex(path); err != nil {
		t.Fatal("companion cannot read upgraded save", err)
	}
}
