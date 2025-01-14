package store

import "fmt"

var (
	ErrAlreadyLocked   error = fmt.Errorf("state already locked")
	ErrAlreadyUnlocked error = fmt.Errorf("state already unlocked")
)

var (
	ErrCreatingLockfile error = fmt.Errorf("failed to create lockfile")
	ErrWritingLockfile  error = fmt.Errorf("failed to write to lockfile")
)
