//go:build !windows

package template

import (
	"fmt"
	"os"
	"syscall"
)

// MkdirAllNoUmask is os.MkdirAll that ignores the current unix umask.
func MkdirAllNoUmask(path string, perms os.FileMode) error {
	existing := syscall.Umask(0)
	defer syscall.Umask(existing)
	err := os.MkdirAll(path, perms)
	if err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	return nil
}
