package migrationstreamcatalog

import (
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"strings"
)

var ownerFilenamePattern = regexp.MustCompile(
	`^[a-z][a-z0-9]*(?:_[a-z][a-z0-9]*)*\.json$`,
)

func validateIncludes(includes []string) (map[string]string, error) {
	if includes == nil {
		return nil, fmt.Errorf("migration stream catalog includes must be an array")
	}
	if len(includes) > maximumIncludes {
		return nil, fmt.Errorf("migration stream catalog include limit exceeded")
	}
	names := make(map[string]string, len(includes))
	previous := ""
	for index, include := range includes {
		if index > 0 && previous == include {
			return nil, fmt.Errorf("duplicate migration stream catalog include %q", include)
		}
		if index > 0 && previous > include {
			return nil, fmt.Errorf("migration stream catalog includes must be sorted")
		}
		name, err := ownerNameFromPath(include)
		if err != nil {
			return nil, err
		}
		names[include] = name
		previous = include
	}
	return names, nil
}

func ownerNameFromPath(include string) (string, error) {
	if !fs.ValidPath(include) || strings.ContainsAny(include, `\:`) ||
		path.Dir(include) != streamDirectoryPath {
		return "", fmt.Errorf("migration stream include path %q is invalid", include)
	}
	filename := path.Base(include)
	if !ownerFilenamePattern.MatchString(filename) {
		return "", fmt.Errorf("migration stream include filename %q is invalid", filename)
	}
	return strings.TrimSuffix(filename, ".json"), nil
}
