//go:build linux || darwin || freebsd

package main

import (
	"fmt"
	"os"
	"syscall"
)

// lockSave holds an advisory lock for the session; the OS releases it on exit.
func lockSave(path string) (func(), error) {
	file, err := os.OpenFile(path+".lock", os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		file.Close()
		return nil, fmt.Errorf("save is already in use or cannot be locked: %w", err)
	}
	return func() { _ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN); _ = file.Close() }, nil
}
