//go:build !linux && !darwin && !freebsd

package main

import "fmt"

func lockSave(path string) (func(), error) {
	return nil, fmt.Errorf("persistent saves currently require Linux, macOS, or FreeBSD")
}
