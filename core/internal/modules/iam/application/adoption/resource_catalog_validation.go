package adoption

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var catalogKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_.:-]{0,127}$`)
var ownerKeyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

func (state *validationState) validateResourceCatalog() {
	resourceMappings := make(map[uuid.UUID]ResourceMapping)
	for _, mapping := range state.mapping.Resources {
		if _, duplicate := resourceMappings[mapping.SourceID]; duplicate {
			state.add(issue("mapping_duplicate", "resources", mapping.SourceID.String()))
		}
		resourceMappings[mapping.SourceID] = mapping
	}
	seen := make(map[string]uuid.UUID)
	for _, source := range state.snapshot.Resources {
		state.validateResource(source, resourceMappings[source.ID], seen)
	}
	for sourceID := range resourceMappings {
		if state.resources[sourceID].ID == uuid.Nil {
			state.add(issue("mapping_source_missing", "resources", sourceID.String()))
		}
	}
}

func (state *validationState) validateResource(
	source SourceResource, mapping ResourceMapping, seen map[string]uuid.UUID,
) {
	key := canonicalCatalogKey(source.Key)
	if source.ID == uuid.Nil || source.TenantID != state.snapshot.TenantID {
		state.add(issue("tenant_or_identity_invalid", "resources", source.ID.String()))
	}
	if _, duplicate := state.resources[source.ID]; duplicate {
		state.add(issue("duplicate_identity", "resources", source.ID.String()))
	}
	state.resources[source.ID] = source
	if prior, duplicate := seen[key]; key == "" || duplicate && prior != source.ID {
		state.add(issue("resource_key_collision", "resources", key))
	}
	seen[key] = source.ID
	if key == "all" {
		if mapping.SourceID != uuid.Nil {
			state.add(issue("wildcard_mapping_unexpected", "resources", source.ID.String()))
		}
		return
	}
	if mapping.SourceID == uuid.Nil || mapping.TargetID == uuid.Nil ||
		mapping.ResourceKey != key || !catalogKeyPattern.MatchString(mapping.ResourceKey) ||
		!ownerKeyPattern.MatchString(mapping.OwnerDomain) ||
		!ownerKeyPattern.MatchString(mapping.ResourceKind) {
		state.add(issue("resource_mapping_invalid", "resources", source.ID.String()))
	}
}

func canonicalCatalogKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
