package chmod

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const setuid = 4
const setgid = 2
const sticky = 1
const read = 4
const write = 2
const exec = 1

const specialMask = 07000
const ownerMask = 0700
const groupMask = 070
const otherMask = 07

func calculateFileMode(curMode os.FileMode, isFolder bool, applyMode string) (os.FileMode, error) {
	// From chmod man page:
	// Each MODE is of the form '[ugoa]*([-+=]([rwxXst]*|[ugo]))+|[-+=][0-7]+'.

	newMode := int64(curMode)

	numericRe := regexp.MustCompile(`^([+=-]?)([0-7]{1,4})$`)
	symbolicRe := regexp.MustCompile(`^([ugoa]*)([+=-]([rwxXst]*|[ugo]))+$`)

	for _, mode := range strings.Split(applyMode, ",") {
		if numericRe.MatchString(mode) {
			parts := numericRe.FindStringSubmatch(mode)
			thisMode, err := strconv.ParseInt(parts[2], 8, 32)
			if err != nil {
				return curMode, err
			}
			modifier := parts[1]
			newMode, err = applyNumericPerms(newMode, modifier, thisMode)
			if err != nil {
				return curMode, nil
			}
		} else if symbolicRe.MatchString(mode) {
			parts := symbolicRe.FindStringSubmatch(mode)
			var err error
			newMode, err = applySymbolicPerms(newMode, isFolder, parts[1], parts[2])
			if err != nil {
				return curMode, err
			}
		} else {
			return curMode, fmt.Errorf("Mode %s is not valid", mode)
		}
	}

	return os.FileMode(newMode), nil
}

func applyNumericPerms(curMode int64, modifier string, perms int64) (int64, error) {
	if modifier == "" {
		modifier = "="
	}

	switch modifier {
	case "=":
		return perms, nil
	case "+":
		return curMode | perms, nil
	case "-":
		return curMode ^ perms, nil
	default:
		return curMode, fmt.Errorf("Unknown permissions modifier [%s]", modifier)
	}
}

func applySymbolicPerms(curMode int64, isFolder bool, appliesTo string, perms string) (int64, error) {
	for _, parts := range regexp.MustCompile(`([+=-])([rwxXst]*|[ugo])`).FindAllStringSubmatch(perms, -1) {
		// modifier := parts[1]
		perm := parts[2]
		switch string(perm) {

		}
	}
	return curMode, nil
}

func getPerms(mode int64, who string) (uint8, error) {
	switch who {
	case "s":
		return uint8((mode & specialMask) >> 32), nil
	case "u":
		return uint8((mode & ownerMask) >> 16), nil
	case "g":
		return uint8((mode & groupMask)) >> 8, nil
	case "o":
		return uint8(mode & otherMask), nil
	default:
		return 0, fmt.Errorf("Invalid source specification [%s]", who)
	}
}

func setPerms(mode int64, modifier string, who string, perms uint8) (int64, error) {
	newMode := int64(perms)
	mask := int64(0)

	switch who {
	case "s":
		newMode <<= 24
		mask = specialMask
	case "u":
		newMode <<= 16
		mask = ownerMask
	case "g":
		newMode <<= 8
		mask = groupMask
	case "o":
		mask = otherMask
	default:
		return mode, fmt.Errorf("Invalid who specification on setPerms [%s]", who)
	}

	switch modifier {
	case "+":
		mode |= newMode
	case "-":
		mode ^= newMode
	case "=":
		mode ^= mask
		mode |= newMode
	}

	return mode, nil
}

func copyPerms(srcMode int64, whoFrom string, whoTo string) (int64, error) {
	perms, err := getPerms(srcMode, whoFrom)
	if err != nil {
		return srcMode, nil
	}
	return setPerms(srcMode, "=", whoTo, perms)
}

// Tests
// +004
// -040
// g-w+r
// g-w+
