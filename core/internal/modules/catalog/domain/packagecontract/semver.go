package packagecontract

import (
	"fmt"
	"regexp"
	"strconv"
)

var semverPattern = regexp.MustCompile(
	`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?$`,
)

type Version struct {
	Major, Minor, Patch uint64
	Prerelease, Build   string
	raw                 string
}

func ParseVersion(raw string) (Version, error) {
	match := semverPattern.FindStringSubmatch(raw)
	if match == nil || !validPrerelease(match[4]) {
		return Version{}, ErrManifestInvalid
	}
	major, err := strconv.ParseUint(match[1], 10, 64)
	if err != nil {
		return Version{}, ErrManifestInvalid
	}
	minor, err := strconv.ParseUint(match[2], 10, 64)
	if err != nil {
		return Version{}, ErrManifestInvalid
	}
	patch, err := strconv.ParseUint(match[3], 10, 64)
	if err != nil {
		return Version{}, ErrManifestInvalid
	}
	return Version{Major: major, Minor: minor, Patch: patch,
		Prerelease: match[4], Build: match[5], raw: raw}, nil
}

func validPrerelease(value string) bool {
	if value == "" {
		return true
	}
	start := 0
	for index := 0; index <= len(value); index++ {
		if index != len(value) && value[index] != '.' {
			continue
		}
		part := value[start:index]
		if part == "" || len(part) > 1 && part[0] == '0' && numeric(part) {
			return false
		}
		start = index + 1
	}
	return true
}

func numeric(value string) bool {
	for index := range value {
		if value[index] < '0' || value[index] > '9' {
			return false
		}
	}
	return true
}

func (value Version) String() string { return value.raw }
func (value Version) Valid() bool    { return value.raw != "" }

func (value Version) Compare(other Version) int {
	for _, pair := range [][2]uint64{{value.Major, other.Major}, {value.Minor, other.Minor}, {value.Patch, other.Patch}} {
		if pair[0] < pair[1] {
			return -1
		}
		if pair[0] > pair[1] {
			return 1
		}
	}
	return comparePrerelease(value.Prerelease, other.Prerelease)
}

func comparePrerelease(left, right string) int {
	if left == right {
		return 0
	}
	if left == "" {
		return 1
	}
	if right == "" {
		return -1
	}
	if left < right {
		return -1
	}
	return 1
}

func (value Version) MarshalText() ([]byte, error) {
	if !value.Valid() {
		return nil, fmt.Errorf("marshal semver: %w", ErrManifestInvalid)
	}
	return []byte(value.raw), nil
}
