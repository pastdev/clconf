package memkv

import (
	"errors"
	"fmt"
	"path"
)

var (
	ErrNotExist   = errors.New("key does not exist")
	ErrBadPattern = path.ErrBadPattern
)

type KeyError struct {
	Key string
	Err error
}

func (e KeyError) Error() string {
	return fmt.Sprintf("%s: %s", e.Key, e.Err.Error())
}

func (e KeyError) Unwrap() error {
	return e.Err
}

func IsBadPattern(err error) bool {
	return errors.Is(err, ErrBadPattern)
}

func IsNotExists(err error) bool {
	return errors.Is(err, ErrNotExist)
}

func NewKeyError(key string, err error) error {
	return &KeyError{Key: key, Err: err}
}
