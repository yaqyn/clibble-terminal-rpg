package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

func cleanInput(text string) []string {
	words := strings.Fields(strings.ToLower(text))
	if words == nil {
		return []string{}
	}
	return words
}

func runREPL(cfg *config, in io.Reader) error {
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 4096), 64*1024)
	for {
		if err := cfg.ctx.Err(); err != nil {
			return err
		}
		if _, err := fmt.Fprint(cfg.out, "Pokedex > "); err != nil {
			return err
		}
		if !scanner.Scan() {
			return scanner.Err()
		}
		words := cleanInput(scanner.Text())
		if len(words) == 0 {
			continue
		}
		cmd, ok := cfg.commands[words[0]]
		if !ok {
			fmt.Fprintln(cfg.out, "Unknown command")
			continue
		}
		args := words[1:]
		if len(args) != cmd.args {
			fmt.Fprintf(cfg.out, "Usage: %s\n", cmd.name)
			continue
		}
		if len(args) > 0 && !validName(args[0]) {
			fmt.Fprintln(cfg.out, "Use a Pokemon or area name containing letters, numbers, and hyphens.")
			continue
		}
		if err := cmd.callback(cfg, args...); errors.Is(err, errExit) {
			return nil
		} else if err != nil {
			if cfg.ctx.Err() != nil {
				return cfg.ctx.Err()
			}
			fmt.Fprintf(cfg.out, "Error: %v\n", err)
		}
	}
}
