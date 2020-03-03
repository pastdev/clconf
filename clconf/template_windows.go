package clconf

import (
	"os"
)

// MkdirAllNoUmask is os.MkdirAll that ignores the current unix umask.
func MkdirAllNoUmask(path string, perms os.FileMode) error {
	// syscall.Umask is not available on windows
	return os.MkdirAll(path, perms)
}
