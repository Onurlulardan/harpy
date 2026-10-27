package adoption

import (
	"strings"

	"github.com/google/uuid"
)

func (state *validationState) validateOrganizations() {
	slugs := make(map[string]uuid.UUID)
	for _, organization := range state.snapshot.Organizations {
		key := organization.ID.String()
		if organization.ID == uuid.Nil || organization.TenantID != state.snapshot.TenantID {
			state.add(issue("tenant_or_identity_invalid", "organizations", key))
		}
		if _, duplicate := state.orgs[organization.ID]; duplicate {
			state.add(issue("duplicate_identity", "organizations", key))
		}
		state.orgs[organization.ID] = organization
		slug := strings.TrimSpace(organization.Slug)
		if prior, duplicate := slugs[slug]; slug == "" || (duplicate && prior != organization.ID) {
			state.add(issue("slug_collision", "organizations", slug))
		}
		slugs[slug] = organization.ID
	}
	parents := make(map[uuid.UUID]uuid.UUID)
	for _, organization := range state.snapshot.Organizations {
		if organization.ParentID == "" {
			continue
		}
		parentID, err := uuid.Parse(organization.ParentID)
		if err != nil || state.orgs[parentID].ID == uuid.Nil {
			state.add(issue("dangling_parent", "organizations", organization.ID.String()))
			continue
		}
		parents[organization.ID] = parentID
	}
	for id := range state.orgs {
		seen := make(map[uuid.UUID]bool)
		for current := id; current != uuid.Nil; current = parents[current] {
			if seen[current] {
				state.add(issue("organization_cycle", "organizations", id.String()))
				break
			}
			seen[current] = true
		}
	}
}
