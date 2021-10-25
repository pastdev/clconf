package template

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pastdev/clconf/v2/clconf/memkv"
)

func NewFuncMap(s *memkv.Store) map[string]interface{} {
	m := make(map[string]interface{})
	m["base"] = path.Base
	m["split"] = strings.Split
	m["json"] = UnmarshalJsonObject
	m["jsonArray"] = UnmarshalJsonArray
	m["dir"] = path.Dir
	m["map"] = CreateMap
	m["getenv"] = Getenv
	m["join"] = strings.Join
	m["datetime"] = time.Now
	m["toUpper"] = strings.ToUpper
	m["toLower"] = strings.ToLower
	m["contains"] = strings.Contains
	m["replace"] = strings.Replace
	m["trimSuffix"] = strings.TrimSuffix
	m["lookupIP"] = LookupIP
	m["lookupIPV4"] = LookupIPV4
	m["lookupIPV6"] = LookupIPV6
	m["lookupSRV"] = LookupSRV
	m["fileExists"] = isFileExist
	m["base64Encode"] = Base64Encode
	m["base64Decode"] = Base64Decode
	m["parseBool"] = strconv.ParseBool
	m["regexReplace"] = RegexReplace
	m["reverse"] = Reverse
	m["sortByLength"] = SortByLength
	m["sortKVByLength"] = SortKVByLength
	m["add"] = func(a, b int) int { return a + b }
	m["sub"] = func(a, b int) int { return a - b }
	m["div"] = func(a, b int) int { return a / b }
	m["mod"] = func(a, b int) int { return a % b }
	m["mul"] = func(a, b int) int { return a * b }
	m["seq"] = Seq
	m["atoi"] = strconv.Atoi
	m["escapeOsgi"] = EscapeOsgi
	m["fqdn"] = Fqdn
	m["sort"] = sortAs
	m["getsvs"] = getsvs(s)
	m["getksvs"] = getksvs(s)
	return m
}

func AddFuncs(out, in map[string]interface{}) {
	for name, fn := range in {
		out[name] = fn
	}
}

// Seq creates a sequence of integers. It's named and used as GNU's seq.
// Seq takes the first and the last element as arguments. So Seq(3, 5) will generate [3,4,5]
func Seq(first, last int) []int {
	var arr []int
	for i := first; i <= last; i++ {
		arr = append(arr, i)
	}
	return arr
}

type byLengthKV []memkv.KVPair

func (s byLengthKV) Len() int {
	return len(s)
}

func (s byLengthKV) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byLengthKV) Less(i, j int) bool {
	return len(s[i].Key) < len(s[j].Key)
}

func SortKVByLength(values []memkv.KVPair) []memkv.KVPair {
	sort.Sort(byLengthKV(values))
	return values
}

type byLength []string

func (s byLength) Len() int {
	return len(s)
}
func (s byLength) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byLength) Less(i, j int) bool {
	return len(s[i]) < len(s[j])
}

func SortByLength(values []string) []string {
	sort.Sort(byLength(values))
	return values
}

//Reverse returns the array in reversed order
//works with []string and []KVPair
func Reverse(values interface{}) interface{} {
	switch v := values.(type) {
	case []string:
		for left, right := 0, len(v)-1; left < right; left, right = left+1, right-1 {
			v[left], v[right] = v[right], v[left]
		}
	case []memkv.KVPair:
		for left, right := 0, len(v)-1; left < right; left, right = left+1, right-1 {
			v[left], v[right] = v[right], v[left]
		}
	}
	return values
}

// Getenv retrieves the value of the environment variable named by the key.
// It returns the value, which will the default value if the variable is not present.
// If no default value was given - returns "".
func Getenv(key string, v ...string) string {
	defaultValue := ""
	if len(v) > 0 {
		defaultValue = v[0]
	}

	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// CreateMap creates a key-value map of string -> interface{}
// The i'th is the key and the i+1 is the value
func CreateMap(values ...interface{}) (map[string]interface{}, error) {
	if len(values)%2 != 0 {
		return nil, errors.New("invalid map call")
	}
	dict := make(map[string]interface{}, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, errors.New("map keys must be strings")
		}
		dict[key] = values[i+1]
	}
	return dict, nil
}

func UnmarshalJsonObject(data string) (map[string]interface{}, error) {
	var ret map[string]interface{}
	err := json.Unmarshal([]byte(data), &ret)
	return ret, err
}

func UnmarshalJsonArray(data string) ([]interface{}, error) {
	var ret []interface{}
	err := json.Unmarshal([]byte(data), &ret)
	return ret, err
}

func LookupIP(data string) []string {
	ips, err := net.LookupIP(data)
	if err != nil {
		return nil
	}
	// "Cast" IPs into strings and sort the array
	ipStrings := make([]string, len(ips))

	for i, ip := range ips {
		ipStrings[i] = ip.String()
	}
	sort.Strings(ipStrings)
	return ipStrings
}

func LookupIPV6(data string) []string {
	var addresses []string
	for _, ip := range LookupIP(data) {
		if strings.Contains(ip, ":") {
			addresses = append(addresses, ip)
		}
	}
	return addresses
}

func LookupIPV4(data string) []string {
	var addresses []string
	for _, ip := range LookupIP(data) {
		if strings.Contains(ip, ".") {
			addresses = append(addresses, ip)
		}
	}
	return addresses
}

type sortSRV []*net.SRV

func (s sortSRV) Len() int {
	return len(s)
}

func (s sortSRV) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortSRV) Less(i, j int) bool {
	str1 := fmt.Sprintf("%s%d%d%d", s[i].Target, s[i].Port, s[i].Priority, s[i].Weight)
	str2 := fmt.Sprintf("%s%d%d%d", s[j].Target, s[j].Port, s[j].Priority, s[j].Weight)
	return str1 < str2
}

func LookupSRV(service, proto, name string) []*net.SRV {
	_, addrs, err := net.LookupSRV(service, proto, name)
	if err != nil {
		return []*net.SRV{}
	}
	sort.Sort(sortSRV(addrs))
	return addrs
}

func Base64Encode(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

func Base64Decode(data string) (string, error) {
	s, err := base64.StdEncoding.DecodeString(data)
	return string(s), err
}

func EscapeOsgi(data string) string {
	// quotes, double quotes, backslash, the equals sign and spaces need to be escaped
	var buffer bytes.Buffer
	for _, runeValue := range data {
		switch runeValue {
		case 39, 34, 92, 61, 32:
			buffer.WriteRune(92)
			buffer.WriteRune(runeValue)
		default:
			buffer.WriteRune(runeValue)
		}
	}
	return buffer.String()
}

// Fqdn returns hostname if it contains a ., otherwise returns hostname.domain
func Fqdn(hostname, domain string) string {
	if strings.Contains(hostname, ".") {
		return hostname
	}
	return hostname + "." + domain
}

// copied from confd util.go
func isFileExist(fpath string) bool {
	if _, err := os.Stat(fpath); os.IsNotExist(err) {
		return false
	}
	return true
}

// RegexReplace maps to regexp.ReplaceAllString
func RegexReplace(regex, src, repl string) (string, error) {
	re, err := regexp.Compile(regex)
	if err != nil {
		return "", err
	}
	return re.ReplaceAllString(src, repl), nil
}

// sortType accepts an array because the input is an optional name of the type
// for sorting and the actual implementation methods (getsvs, getksvs) accept
// varargs so this utility function allows direct call without unpacking.
func sortType(input []string) (string, error) {
	r := "string"
	if len(input) > 0 {
		r = input[0]
	}
	if r != "string" && r != "int" {
		return "", fmt.Errorf("sort: Type '%s' is not supported (only int, string)", r)
	}
	return r, nil
}

type asInt []string

func (p asInt) Len() int { return len(p) }
func (p asInt) Less(i, j int) bool {
	a, aerr := strconv.Atoi(p[i])
	b, berr := strconv.Atoi(p[j])

	if aerr == nil {
		if berr == nil {
			return a < b
		}
		return true // Numbers come first
	} else {
		if berr == nil {
			return false
		}
	}
	return p[i] < p[j]
}
func (p asInt) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// sortAs sorts the input in-place as specified type (string, int, default: string) and returns it.
func sortAs(v []string, asType ...string) ([]string, error) {
	sortType, err := sortType(asType)
	if err != nil {
		return nil, err
	}

	if sortType == "int" {
		sort.Sort(asInt(v))
	} else {
		sort.Strings(v)
	}
	return v, nil
}

func getsvs(s *memkv.Store) func(string, ...string) ([]string, error) {
	return func(pattern string, asType ...string) ([]string, error) {
		vals, err := s.GetAllValues(pattern)
		if err != nil {
			return nil, err
		}
		return sortAs(vals, asType...)
	}
}

func getksvs(s *memkv.Store) func(string, ...string) ([]string, error) {
	return func(pattern string, asType ...string) ([]string, error) {
		ks, err := s.GetAll(pattern)
		if err != nil {
			return nil, err
		}
		kvMap := make(map[string]string)
		keys := make([]string, len(ks))
		r := make([]string, len(ks))
		for i, kv := range ks {
			key := kv.Key[strings.LastIndex(kv.Key, "/")+1:]
			kvMap[key] = kv.Value
			keys[i] = key
		}

		keys, err = sortAs(keys, asType...)
		if err != nil {
			return nil, err
		}
		for i, key := range keys {
			r[i] = kvMap[key]
		}
		return r, nil
	}
}
