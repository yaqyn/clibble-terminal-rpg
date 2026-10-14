package main

import (
	"fmt"

	"github.com/gofrs/flock"
)

// lockSave holds an exclusive advisory lock until the returned function runs.
// OS file locks are released automatically when the process exits.
func lockSave(path string) (func(), error) {
	lock := flock.New(path+".lock", flock.SetPermissions(0600))
	locked, err := lock.TryLock()
	if err != nil {
		_ = lock.Close()
		return nil, fmt.Errorf("lock save: %w", err)
	}
	if !locked {
		_ = lock.Close()
		return nil, fmt.Errorf("save is already in use by another Dexrove session")
	}
	return func() { _ = lock.Close() }, nil
}
