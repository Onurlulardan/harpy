package packagecontract

import "strings"

type Range struct {
	kind  byte
	lower Version
	raw   string
}

func ParseRange(raw string) (Range, error) {
	if strings.HasPrefix(raw, "=") {
		version, err := ParseVersion(strings.TrimPrefix(raw, "="))
		if err != nil || version.Prerelease != "" {
			return Range{}, ErrManifestInvalid
		}
		return Range{kind: '=', lower: version, raw: raw}, nil
	}
	if strings.HasPrefix(raw, "^") {
		version, err := ParseVersion(strings.TrimPrefix(raw, "^"))
		if err != nil || version.Prerelease != "" {
			return Range{}, ErrManifestInvalid
		}
		return Range{kind: '^', lower: version, raw: raw}, nil
	}
	return Range{}, ErrManifestInvalid
}

func (value Range) Contains(candidate Version) bool {
	if !value.Valid() || !candidate.Valid() || candidate.Prerelease != "" {
		return false
	}
	if value.kind == '=' {
		return value.lower.Compare(candidate) == 0
	}
	if value.lower.Compare(candidate) > 0 {
		return false
	}
	if value.lower.Major > 0 {
		return candidate.Major == value.lower.Major
	}
	if value.lower.Minor > 0 {
		return candidate.Major == 0 && candidate.Minor == value.lower.Minor
	}
	return candidate.Major == 0 && candidate.Minor == 0 &&
		candidate.Patch == value.lower.Patch
}

func (value Range) String() string { return value.raw }
func (value Range) Valid() bool    { return value.raw != "" }
