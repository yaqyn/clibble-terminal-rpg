package main

import (
	"bytes"
	"testing"
)

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
