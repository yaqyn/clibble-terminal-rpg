package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestTrailsListsBundledGuideWithoutAnAPI(t *testing.T) {
	var output bytes.Buffer
	if err := commandTrails(&config{out: &output}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "mosslight-woods: 3 wildlings") ||
		!strings.Contains(output.String(), "old-root-ridge: 3 wildlings") {
		t.Fatal(output.String())
	}
}

func TestWhereListsBundledHabitatsWithoutAnAPI(t *testing.T) {
	var output bytes.Buffer
	cfg := &config{out: &output}
	if err := commandWhere(cfg, "mosskin"); err != nil {
		t.Fatal(err)
	}
	if output.String() != "mosslight-woods\nbramble-hollow\n" {
		t.Fatal(output.String())
	}
	output.Reset()
	if err := commandWhere(cfg, "unknown"); err == nil {
		t.Fatal("unknown wildling accepted")
	}
	if output.Len() != 0 {
		t.Fatal("unknown wildling produced misleading locations")
	}
}
