package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"golang.org/x/term"
)

type terminalIO struct {
	io.Reader
	io.Writer
}

// enableEditor adds native line editing only to an actual interactive terminal.
// History stays in memory; no command log is written to disk.
func enableEditor(cfg *config, in io.Reader, out io.Writer) (func(), error) {
	input, inOK := in.(*os.File)
	output, outOK := out.(*os.File)
	if !inOK || !outOK || !term.IsTerminal(int(input.Fd())) || !term.IsTerminal(int(output.Fd())) || os.Getenv("TERM") == "dumb" {
		return func() {}, nil
	}
	state, err := term.MakeRaw(int(input.Fd()))
	if err != nil {
		return nil, fmt.Errorf("enable terminal editing: %w", err)
	}
	ctx, cancel := context.WithCancel(cfg.ctx)
	cfg.ctx = ctx
	editor := term.NewTerminal(terminalIO{input, output}, cfg.view.accent("dexrove")+" > ")
	if width, height, err := term.GetSize(int(output.Fd())); err == nil {
		_ = editor.SetSize(width, height)
	}
	commands := make([]string, 0, len(cfg.commands))
	for key := range cfg.commands {
		commands = append(commands, key)
	}
	sort.Strings(commands)
	editor.AutoCompleteCallback = func(line string, pos int, key rune) (string, int, bool) {
		if key == 3 {
			cancel()
			return "", 0, true
		}
		replacement, cursor, ok := completeCommand(commands, line, pos, key)
		return replacement, cursor, ok
	}
	cfg.view.editing = true
	cfg.readLine = editor.ReadLine
	cfg.prompt = false
	cfg.out = editor
	cfg.errOut = editor
	cfg.view.out = editor
	return func() { cancel(); _ = term.Restore(int(input.Fd()), state); fmt.Fprintln(output) }, nil
}
func completeCommand(commands []string, line string, pos int, key rune) (string, int, bool) {
	if key != '\t' || pos != len(line) || strings.ContainsAny(line, " \t") || line == "" {
		return line, pos, false
	}
	matches := make([]string, 0)
	for _, command := range commands {
		if strings.HasPrefix(command, strings.ToLower(line)) {
			matches = append(matches, command)
		}
	}
	if len(matches) == 0 {
		return line, pos, false
	}
	result := matches[0]
	if len(matches) == 1 {
		result += " "
		return result, len(result), true
	}
	for _, match := range matches[1:] {
		for !strings.HasPrefix(match, result) {
			result = result[:len(result)-1]
		}
	}
	return result, len(result), true
}
