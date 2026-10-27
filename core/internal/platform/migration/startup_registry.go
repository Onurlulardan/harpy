package migration

import "errors"

var errHarpyFreshSchemaAbsent = errors.New("Harpy fresh schema is not present")

func LoadStartupRegistry(documents []Document) (*ReleaseRegistry, error) {
	selected, err := SelectFreshDocuments(documents)
	if err != nil {
		return nil, err
	}
	if len(selected) == 0 {
		return nil, errHarpyFreshSchemaAbsent
	}
	return newFreshRegistry(selected)
}

func (registry *ReleaseRegistry) FreshEntryCount() int {
	if registry == nil {
		return 0
	}
	return len(registry.freshPaths)
}

func (registry *ReleaseRegistry) FreshPaths() []string {
	if registry == nil {
		return nil
	}
	return append([]string(nil), registry.freshPaths...)
}
