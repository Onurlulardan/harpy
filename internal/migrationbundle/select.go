package migrationbundle

import (
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"
)

type freshKey struct {
	owner   string
	scope   string
	version int64
}

type freshCandidate struct {
	source  Source
	version int64
	rank    int
}

func Select(sources []Source) ([]Source, error) {
	if len(sources) == 0 {
		return nil, fmt.Errorf("fresh migration sources are absent")
	}
	best := make(map[freshKey]freshCandidate, len(sources))
	seenPath := make(map[string]struct{}, len(sources))
	for _, source := range sources {
		candidate, err := freshCandidateFrom(source)
		if err != nil {
			return nil, err
		}
		if _, duplicate := seenPath[source.Path]; duplicate {
			return nil, fmt.Errorf("fresh migration %s is duplicated", source.Path)
		}
		seenPath[source.Path] = struct{}{}
		key := freshKey{owner: source.Owner, scope: source.Scope, version: candidate.version}
		current, exists := best[key]
		if !exists || candidate.rank > current.rank {
			best[key] = candidate
			continue
		}
		if candidate.rank == current.rank {
			return nil, fmt.Errorf(
				"fresh migration version %d has ambiguous supersession", candidate.version,
			)
		}
	}
	selected := make([]freshCandidate, 0, len(best))
	for _, candidate := range best {
		selected = append(selected, candidate)
	}
	sort.Slice(selected, func(left, right int) bool {
		if selected[left].version != selected[right].version {
			return selected[left].version < selected[right].version
		}
		return selected[left].source.Path < selected[right].source.Path
	})
	sources = make([]Source, len(selected))
	for index, candidate := range selected {
		sources[index] = candidate.source
	}
	return sources, nil
}

func freshCandidateFrom(source Source) (freshCandidate, error) {
	if strings.Contains(source.Path, "/v1_adopted/") ||
		strings.HasSuffix(source.Path, ".v1_adopted.sql") {
		return freshCandidate{}, fmt.Errorf("v1_adopted is not present: %s", source.Path)
	}
	version, err := freshVersion(source.Path)
	if err != nil {
		return freshCandidate{}, err
	}
	rank, err := supersessionRank(source.Path)
	if err != nil {
		return freshCandidate{}, err
	}
	return freshCandidate{source: source, version: version, rank: rank}, nil
}

func freshVersion(documentPath string) (int64, error) {
	base := path.Base(documentPath)
	prefix, _, ok := strings.Cut(base, "_")
	if !ok || len(prefix) != 14 {
		return 0, fmt.Errorf("fresh migration version is invalid: %s", documentPath)
	}
	version, err := strconv.ParseInt(prefix, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("fresh migration version is invalid: %s", documentPath)
	}
	return version, nil
}

func supersessionRank(documentPath string) (int, error) {
	const marker = "/supersessions/"
	index := strings.Index(documentPath, marker)
	if index < 0 {
		return 0, nil
	}
	rest := documentPath[index+len(marker):]
	level, _, ok := strings.Cut(rest, "/")
	if !ok {
		return 0, fmt.Errorf("fresh migration supersession path is invalid: %s", documentPath)
	}
	if !strings.HasPrefix(level, "v") {
		return 1, nil
	}
	rank, err := strconv.Atoi(level[1:])
	if err != nil || rank < 1 {
		return 0, fmt.Errorf("fresh migration supersession path is invalid: %s", documentPath)
	}
	return rank, nil
}
