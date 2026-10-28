package storagecatalog

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

var fragmentFilenamePattern = regexp.MustCompile(
	`^[0-9]{2}_[a-z][a-z0-9]*(?:_[a-z0-9]+)*\.json$`,
)

func validateIncludes(includes []string) (map[string]string, error) {
	if includes == nil {
		return nil, fmt.Errorf("storage catalog includes must be an array")
	}
	if len(includes) > maximumIncludes {
		return nil, fmt.Errorf("storage catalog include limit exceeded")
	}
	names := make(map[string]string, len(includes))
	previous := ""
	for index, include := range includes {
		if index > 0 && previous == include {
			return nil, fmt.Errorf("duplicate storage catalog include %q", include)
		}
		if index > 0 && previous > include {
			return nil, fmt.Errorf("storage catalog includes must be sorted")
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
		path.Dir(include) != ownerDirectoryPath {
		return "", fmt.Errorf("storage catalog include path %q is invalid", include)
	}
	filename := path.Base(include)
	if !ownerFilenamePattern.MatchString(filename) {
		return "", fmt.Errorf("storage catalog include filename %q is invalid", filename)
	}
	return strings.TrimSuffix(filename, ".json"), nil
}

func validateOwnerFragmentIncludes(owner string, includes []string) ([]string, error) {
	if includes == nil || len(includes) == 0 {
		return nil, fmt.Errorf("modular storage owner %q includes must not be empty", owner)
	}
	if len(includes) > maximumIncludes {
		return nil, fmt.Errorf("storage owner fragment include limit exceeded")
	}
	previous := ""
	for index, include := range includes {
		if index > 0 && previous == include {
			return nil, fmt.Errorf("duplicate storage owner fragment %q", include)
		}
		if index > 0 && previous > include {
			return nil, fmt.Errorf("storage owner fragment includes must be sorted")
		}
		if err := validateOwnerFragmentPath(owner, include); err != nil {
			return nil, err
		}
		previous = include
	}
	return includes, nil
}

func validateOwnerFragmentPath(owner, include string) error {
	if !fs.ValidPath(include) || strings.ContainsAny(include, `\:`) {
		return fmt.Errorf("storage owner fragment path %q is invalid", include)
	}
	directory := fragmentDirectoryPath + "/" + owner
	if path.Dir(include) != directory || !fragmentFilenamePattern.MatchString(path.Base(include)) {
		return fmt.Errorf("storage owner fragment path %q is invalid", include)
	}
	return nil
}
