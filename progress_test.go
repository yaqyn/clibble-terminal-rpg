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
	cfg.saveProgress = func(p journalRecords, g gameProgress) error { return saveGame(path, p, g) }
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
	if cfg.progress.Seeds != 9 {
		t.Fatal(cfg.progress)
	}
	if err := commandCatch(cfg, "pidgey"); err == nil {
		t.Fatal("fourth throw permitted")
	}
	now = now.Add(encounterWindow)
	if err := commandCatch(cfg, "pidgey"); err != nil {
		t.Fatal(err)
	}
	if cfg.progress.Seeds != 11 || cfg.progress.Encounters["pidgey"].Offers != 1 {
		t.Fatal(cfg.progress)
	}
}
func TestCatchBudgetRollsBackOnFailure(t *testing.T) {
	cfg, _, _ := testConfig()
	now := time.Now()
	cfg.progress = freshProgress(now)
	cfg.saveProgress = func(journalRecords, gameProgress) error { return errors.New("disk full") }
	_, err := attemptCatch(context.Background(), cfg.client, cfg.pokedex, func() float64 { return 1 }, cfg.persist, cfg.progress, now, "pidgey")
	if err == nil || cfg.progress.Seeds != 12 || len(cfg.progress.Encounters) != 0 {
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
	if g.Seeds != 12 {
		t.Fatal(g)
	}
	g.Seeds = 4
	g.MotionOff = true
	if err := saveGame(path, p, g); err != nil {
		t.Fatal(err)
	}
	_, g, err = loadGame(path)
	if err != nil || g.Seeds != 4 || !g.MotionOff {
		t.Fatal(g, err)
	}
	if _, err := loadPokedex(path); err != nil {
		t.Fatal("companion cannot read upgraded save", err)
	}
}

func TestSupplyRefillCapAndClockRollback(t *testing.T) {
	now := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	p := freshProgress(now)
	p.Seeds = 0
	if got := p.refreshed(now.Add(4 * time.Minute)); got.Seeds != 0 {
		t.Fatal("refilled early")
	}
	if got := p.refreshed(now.Add(11 * time.Minute)); got.Seeds != 2 || !got.RefillAt.Equal(now.Add(10*time.Minute)) {
		t.Fatal(got)
	}
	if got := p.refreshed(now.Add(24 * time.Hour)); got.Seeds != 12 {
		t.Fatal("capacity exceeded")
	}
	if got := p.refreshed(now.Add(-time.Hour)); got.Seeds != 0 {
		t.Fatal("clock rollback granted supplies")
	}
}
func TestDataFailureDoesNotSpendOrPersist(t *testing.T) {
	cfg, _, api := testConfig()
	api.err = errors.New("unavailable")
	cfg.progress = freshProgress(time.Now())
	cfg.saveProgress = func(journalRecords, gameProgress) error { t.Fatal("failed lookup saved a budget"); return nil }
	if err := commandCatch(cfg, "mosskin"); err == nil {
		t.Fatal("lookup failure hidden")
	}
	if cfg.progress.Seeds != 12 {
		t.Fatal("failed lookup spent a seed")
	}
}
