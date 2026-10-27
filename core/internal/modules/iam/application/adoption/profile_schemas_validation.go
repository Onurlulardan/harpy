package adoption

import (
	"sort"

	"github.com/google/uuid"
)

var profileKinds = map[string]bool{
	"user": true, "role": true, "organization": true,
}

func (state *validationState) validateProfileSchemas() {
	wanted := make(map[string]bool)
	for _, field := range state.snapshot.ProfileFields {
		wanted[canonicalCatalogKey(field.ResourceKind)] = true
	}
	seen := make(map[string]bool)
	for _, mapping := range state.mapping.ProfileSchemas {
		kind := canonicalCatalogKey(mapping.ResourceKind)
		if seen[kind] {
			state.add(issue("mapping_duplicate", "profiles", kind))
		}
		seen[kind] = true
		if !wanted[kind] || !profileKinds[kind] || mapping.SchemaVersionID == uuid.Nil {
			state.add(issue("profile_schema_mapping_invalid", "profiles", kind))
		}
	}
	for kind := range wanted {
		if !seen[kind] {
			state.add(issue("profile_schema_mapping_missing", "profiles", kind))
		}
	}
}

func sortedProfileKinds(values map[string]bool) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
