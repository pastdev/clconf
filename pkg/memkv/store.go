// Package memkv implements an in-memory key/value store with same API surface
// as "github.com/kelseyhightower/memkv".
package memkv

import (
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"
)

type Option func(s *Store)

type Store struct {
	FuncMap map[string]interface{}
	kv      map[string]string
}

func New(opts ...Option) Store {
	s := Store{kv: map[string]string{}}
	for _, opt := range opts {
		opt(&s)
	}
	s.FuncMap = map[string]interface{}{
		"exists": s.Exists,
		"ls":     s.List,
		"lsdir":  s.ListDir,
		"get":    s.Get,
		"gets":   s.GetAll,
		"getv":   s.GetValue,
		"getvs":  s.GetAllValues,
	}
	return s
}

func (s Store) Del(key string) {
	delete(s.kv, key)
}

func (s Store) Exists(key string) bool {
	_, ok := s.kv[key]
	return ok
}

func (s Store) Get(key string) (KVPair, error) {
	result := KVPair{Key: key}
	var ok bool
	result.Value, ok = s.kv[key]
	if !ok {
		return result, NewKeyError(key, ErrNotExist)
	}
	return result, nil
}

func (s Store) GetAll(pattern string) (KVPairs, error) {
	result := KVPairs{}
	for k, v := range s.kv {
		matches, err := path.Match(pattern, k)
		if err != nil {
			return result, NewKeyError(pattern, ErrBadPattern)
		}
		if matches {
			result = append(result, KVPair{Key: k, Value: v})
		}
	}
	sort.Sort(result)
	return result, nil
}

func (s Store) GetAllValues(pattern string) ([]string, error) {
	result := []string{}
	for k, v := range s.kv {
		matches, err := path.Match(pattern, k)
		if err != nil {
			return result, NewKeyError(pattern, ErrBadPattern)
		}
		if matches {
			result = append(result, v)
		}
	}
	sort.Strings(result)
	return result, nil
}

func (s Store) GetValue(key string, defaultValue ...string) (string, error) {
	v, ok := s.kv[key]
	if !ok {
		if len(defaultValue) == 1 {
			return defaultValue[0], nil
		}
		return "", NewKeyError(key, ErrNotExist)
	}

	return v, nil
}

// list implements path listings similar to the way unix ls command works
// see the template.md for more detail
func (s Store) list(filePath string, dir bool) []string {
	filePath = strings.TrimSuffix(filePath, "/")

	result := []string{}
	seen := map[string]interface{}{}
	for k := range s.kv {
		fullPathMatch := false
		switch {
		case filePath == "":
			// matches all entries
		case strings.HasPrefix(k, filePath):
			switch {
			case len(k) == len(filePath):
				fullPathMatch = true
				// exact match
			case k[len(filePath)] == '/':
				// has subpath
			default:
				// name prefix, but not full name
				continue
			}
		default:
			continue
		}

		var p string
		switch {
		case fullPathMatch:
			if dir {
				continue
			}
			p = k[strings.LastIndex(k, "/")+1:]
		default:
			p = k[len(filePath)+1:]
			if lastSlash := strings.Index(p, "/"); lastSlash > 0 {
				p = p[:lastSlash]
			} else if dir {
				continue
			}
		}

		if _, ok := seen[p]; !ok {
			seen[p] = nil
			result = append(result, p)
		}
	}

	sort.Strings(result)

	return result
}

func (s Store) List(filePath string) []string {
	return s.list(filePath, false)
}

func (s Store) ListDir(filePath string) []string {
	return s.list(filePath, true)
}

func (s Store) Purge() {
	for k := range s.kv {
		delete(s.kv, k)
	}
}

func (s Store) Set(key string, value string) {
	s.kv[key] = value
}

// ToKvMap will return a one-level map of key value pairs where the key is
// a / separated path of subkeys.
func (s Store) ToKvMap() map[string]string {
	result := make(map[string]string, len(s.kv))
	for k, v := range s.kv {
		result[k] = v
	}
	return result
}

// FillKvMap will fill the supplied kvMap with values from data.
func FillKvMap(kvMap map[string]string, data interface{}) {
	Walk(func(keyStack []string, value interface{}) {
		key := "/" + strings.Join(keyStack, "/")
		if value == nil {
			kvMap[key] = ""
		} else {
			kvMap[key] = fmt.Sprintf("%v", value)
		}
	}, data)
}

// Walk will recursively iterate over all the nodes of data calling callback
// for each node.
func Walk(callback func(key []string, value interface{}), data interface{}) {
	walk(callback, data, []string{})
}

func walk(callback func(key []string, value interface{}), node interface{}, keyStack []string) {
	switch typed := node.(type) {
	case map[string]interface{}:
		// json deserialized
		for k, v := range typed {
			keyStack := append(keyStack, fmt.Sprintf("%v", k))
			walk(callback, v, keyStack)
		}
	case map[interface{}]interface{}:
		// yaml deserialized
		for k, v := range typed {
			keyStack := append(keyStack, fmt.Sprintf("%v", k))
			walk(callback, v, keyStack)
		}
	case []interface{}:
		for i, j := range typed {
			keyStack := append(keyStack, strconv.Itoa(i))
			walk(callback, j, keyStack)
		}
	default:
		callback(keyStack, node)
	}
}

func WithKvMap(kv map[string]string) Option {
	return func(s *Store) {
		for k, v := range kv {
			s.kv[k] = v
		}
	}
}

func WithMap(data interface{}) Option {
	return func(s *Store) {
		FillKvMap(s.kv, data)
	}
}
