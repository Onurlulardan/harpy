package main

import (
	"fmt"
	"strconv"
	"strings"
)

type goRelease struct {
	major int
	minor int
	patch int
}

func parseRelease(raw string) (goRelease, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return goRelease{}, fmt.Errorf("version must be a patch release such as 1.27.3")
	}
	numbers := [3]int{}
	for index, part := range parts {
		if part == "" || (len(part) > 1 && part[0] == '0') {
			return goRelease{}, fmt.Errorf("version must be a patch release such as 1.27.3")
		}
		number, err := strconv.Atoi(part)
		if err != nil || number < 0 || (index == 0 && number < 1) {
			return goRelease{}, fmt.Errorf("version must be a patch release such as 1.27.3")
		}
		numbers[index] = number
	}
	release := goRelease{major: numbers[0], minor: numbers[1], patch: numbers[2]}
	if release.String() != raw {
		return goRelease{}, fmt.Errorf("version must be a patch release such as 1.27.3")
	}
	return release, nil
}

func (release goRelease) String() string {
	return fmt.Sprintf("%d.%d.%d", release.major, release.minor, release.patch)
}

func (release goRelease) Toolchain() string {
	return "go" + release.String()
}

func (release goRelease) Language() string {
	return fmt.Sprintf("%d.%d.0", release.major, release.minor)
}

func (release goRelease) Minor() string {
	return fmt.Sprintf("%d.%d", release.major, release.minor)
}

func sameRelease(left, right goRelease) bool {
	return left == right
}
