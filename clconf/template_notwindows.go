// +build !windows

package clconf

import (
	"os"
	"syscall"
)

// MkdirAllNoUmask is os.MkdirAll that ignores the current unix umask.
func MkdirAllNoUmask(path string, perms os.FileMode) error {
	existing := syscall.Umask(0)
	defer syscall.Umask(existing)
	return os.MkdirAll(path, perms)
}
