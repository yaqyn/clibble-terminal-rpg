package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

var errCommandsFailed = errors.New("one or more commands failed")

func cleanInput(text string) []string {
	words := strings.Fields(strings.ToLower(text))
	if words == nil {
		return []string{}
	}
	return words
}

func execute(cfg *config, words []string) error {
	if len(words) == 0 {
		return nil
	}
	cmd, ok := cfg.commands[words[0]]
	if !ok {
		return fmt.Errorf("Unknown command: %q. Use help to see available commands.", words[0])
	}
	args := words[1:]
	if len(args) != cmd.args {
		return fmt.Errorf("Usage: %s", cmd.name)
	}
	if len(args) > 0 && !validName(args[0]) {
		return fmt.Errorf("use a Pokemon or area name containing letters, numbers, and hyphens")
	}
	return cmd.callback(cfg, args...)
}

func runREPL(cfg *config, in io.Reader) error {
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 4096), 64*1024)
	// Terminal reads can remain blocked after Close. Keep input scanning separate
	// so cancellation does not wait for the next newline.
	type inputLine struct {
		text string
		err  error
		done bool
	}
	lines := make(chan inputLine)
	stopScan := make(chan struct{})
	defer close(stopScan)

	readLine := cfg.readLine
	if readLine == nil {
		readLine = func() (string, error) {
			if scanner.Scan() {
				return scanner.Text(), nil
			}
			if err := scanner.Err(); err != nil {
				return "", err
			}
			return "", io.EOF
		}
	}
	go func() {
		for {
			text, err := readLine()
			if err != nil {
				if errors.Is(err, io.EOF) {
					err = nil
				}
				select {
				case lines <- inputLine{done: true, err: err}:
				case <-stopScan:
				}
				return
			}
			select {
			case lines <- inputLine{text: text}:
			case <-stopScan:
				return
			}
		}
	}()

	failed := false
	finish := func(err error) error {
		if err != nil {
			return err
		}
		if cfg.batch && failed {
			return errCommandsFailed
		}
		return nil
	}
	errOut := cfg.errOut
	if errOut == nil {
		errOut = cfg.out
	}
	for {
		if err := cfg.ctx.Err(); err != nil {
			return err
		}
		if cfg.prompt {
			if _, err := fmt.Fprint(cfg.out, "clibble > "); err != nil {
				return err
			}
		}
		var line inputLine
		select {
		case <-cfg.ctx.Done():
			return cfg.ctx.Err()
		case line = <-lines:
		}
		if line.done {
			return finish(line.err)
		}
		words := cleanInput(line.text)
		if len(words) == 0 {
			continue
		}
		if err := execute(cfg, words); errors.Is(err, errExit) {
			return finish(nil)
		} else if err != nil {
			if cfg.ctx.Err() != nil {
				return cfg.ctx.Err()
			}
			failed = true
			if _, writeErr := fmt.Fprintf(errOut, "Error: %v\n", err); writeErr != nil {
				return writeErr
			}
		}
	}
}
