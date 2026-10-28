package migrationbundle

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
)

func Collect(root string) ([]Source, error) {
	var documents []Source
	for _, set := range corpusSets() {
		owned, err := readSet(root, set)
		if err != nil {
			return nil, err
		}
		documents = append(documents, owned...)
	}
	if len(documents) == 0 {
		return nil, fmt.Errorf("fresh migration sources are absent")
	}
	return documents, nil
}

func readSet(root string, set corpusSet) ([]Source, error) {
	base := filepath.Join(root, filepath.FromSlash(set.root))
	seen := map[string]struct{}{}
	documents := make([]Source, 0)
	for _, glob := range set.globs {
		matches, err := filepath.Glob(filepath.Join(base, filepath.FromSlash(glob)))
		if err != nil {
			return nil, err
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("fresh migration glob %s/%s is empty", set.root, glob)
		}
		for _, match := range matches {
			rel, err := filepath.Rel(base, match)
			if err != nil {
				return nil, err
			}
			logical := path.Join(set.root, filepath.ToSlash(rel))
			if _, exists := seen[logical]; exists {
				continue
			}
			seen[logical] = struct{}{}
			body, err := os.ReadFile(match)
			if err != nil {
				return nil, err
			}
			documents = append(documents, Source{
				Owner: set.owner, Scope: set.scope, Path: logical, Body: body,
			})
		}
	}
	if !set.supersede {
		return documents, nil
	}
	return dropSuperseded(set, documents, filepath.Join(base, "migration-supersessions.json"))
}
