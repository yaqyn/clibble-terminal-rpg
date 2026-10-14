//go:build !windows

package main

import "os"

func replaceSave(source, target string) error { return os.Rename(source, target) }
