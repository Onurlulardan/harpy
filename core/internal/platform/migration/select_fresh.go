package migration

import (
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"
)

func SelectFreshDocuments(documents []Document) ([]Document, error) {
	if len(documents) == 0 {
		return nil, errHarpyFreshSchemaAbsent
	}
	best := map[string]Document{}
	bestRank := map[string]int{}
	seen := map[string]struct{}{}
	for _, document := range documents {
		if strings.Contains(document.Path(), "/v1_adopted/") ||
			strings.HasSuffix(document.Path(), ".v1_adopted.sql") {
			return nil, fmt.Errorf("v1_adopted is not present: %s", document.Path())
		}
		version, err := freshDocumentVersion(document.Path())
		if err != nil {
			return nil, err
		}
		rank, err := freshSupersessionRank(document.Path())
		if err != nil {
			return nil, err
		}
		if _, duplicate := seen[document.Path()]; duplicate {
			return nil, fmt.Errorf("fresh migration %s is duplicated", document.Path())
		}
		seen[document.Path()] = struct{}{}
		key := document.Owner() + "\x00" + document.Scope() + "\x00" + strconv.FormatInt(version, 10)
		current, exists := bestRank[key]
		if !exists || rank > current {
			best[key] = document
			bestRank[key] = rank
			continue
		}
		if rank == current {
			return nil, fmt.Errorf("fresh migration version %d has ambiguous supersession", version)
		}
	}
	selected := make([]Document, 0, len(best))
	for _, document := range best {
		selected = append(selected, document)
	}
	sort.Slice(selected, func(left, right int) bool {
		leftVersion, _ := freshDocumentVersion(selected[left].Path())
		rightVersion, _ := freshDocumentVersion(selected[right].Path())
		if leftVersion != rightVersion {
			return leftVersion < rightVersion
		}
		return selected[left].Path() < selected[right].Path()
	})
	return selected, nil
}

func freshDocumentVersion(documentPath string) (int64, error) {
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

func freshSupersessionRank(documentPath string) (int, error) {
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
	if level == "fresh" {
		return 1, nil
	}
	if !strings.HasPrefix(level, "v") {
		return 0, fmt.Errorf("fresh migration supersession path is invalid: %s", documentPath)
	}
	rank, err := strconv.Atoi(level[1:])
	if err != nil || rank < 2 {
		return 0, fmt.Errorf("fresh migration supersession path is invalid: %s", documentPath)
	}
	return rank, nil
}
