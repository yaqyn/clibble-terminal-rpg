package main

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestSuppliesRefreshesWithoutSpendingOrSaving(t *testing.T) {
	now := time.Date(2026, 10, 16, 0, 0, 0, 0, time.UTC)
	progress := freshProgress(now.Add(-11 * time.Minute))
	progress.Seeds = 3
	var output bytes.Buffer
	cfg := &config{progress: progress, clock: func() time.Time { return now }, out: &output}
	if err := commandSupplies(cfg); err != nil {
		t.Fatal(err)
	}
	if output.String() != "Catch seeds: 5/12\nNext seed in: 4m0s\n" {
		t.Fatal(output.String())
	}
	if cfg.progress.Seeds != 3 {
		t.Fatal("inspection changed saved progress")
	}
	output.Reset()
	cfg.progress = freshProgress(now)
	if err := commandSupplies(cfg); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "Next seed") {
		t.Fatal(output.String())
	}
}
