package migration

import (
	"fmt"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
)

func partitionCatalog(
	bound map[string][]byte,
	selected []postgresv1.SelectedMigration,
) (
	map[string]*memoryFS,
	map[string][]postgresv1.SelectedMigration,
	error,
) {
	sourceByScope := map[string]map[string][]byte{}
	selectedByScope := map[string][]postgresv1.SelectedMigration{}
	consumed := make(map[string]bool, len(selected))
	for _, candidate := range selected {
		if candidate.Scope != "platform" && candidate.Scope != "tenant" {
			return nil, nil, fmt.Errorf("selected migration scope is invalid")
		}
		content, exists := bound[candidate.Path]
		if !exists || consumed[candidate.Path] {
			return nil, nil, fmt.Errorf("selected migration source binding is invalid")
		}
		consumed[candidate.Path] = true
		if sourceByScope[candidate.Scope] == nil {
			sourceByScope[candidate.Scope] = map[string][]byte{}
		}
		sourceByScope[candidate.Scope][candidate.Path] = append([]byte(nil), content...)
		selectedByScope[candidate.Scope] = append(
			selectedByScope[candidate.Scope], candidate,
		)
	}
	if len(consumed) != len(bound) {
		return nil, nil, fmt.Errorf("bound migration source set contains an unselected entry")
	}
	filesByScope := make(map[string]*memoryFS, len(sourceByScope))
	for scope, documents := range sourceByScope {
		flattened, err := flattenDocuments(documents)
		if err != nil {
			return nil, nil, err
		}
		filesByScope[scope] = newMemoryFS(flattened)
	}
	return filesByScope, selectedByScope, nil
}
