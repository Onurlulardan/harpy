package application

import (
	fieldregistry "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/fieldregistry"
	presentation "github.com/Onurlulardan/harpy/core/internal/platform/fieldcatalog/presentation"
	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

func publishedVerificationCatalog(
	published PublishedPresentation,
	resource layoutapi.ResourceKind,
	current presentation.FieldCatalog,
) (presentation.FieldCatalog, error) {
	currentJSON, currentSHA, err := catalogEvidence(current)
	if err != nil {
		return presentation.FieldCatalog{}, layoutapi.ErrUnavailable
	}
	if sameBytes(published.CatalogJSON, currentJSON) &&
		sameDigest(published.CatalogSHA256, currentSHA) {
		return current, nil
	}
	if resource != layoutapi.ResourceOrganization {
		return presentation.FieldCatalog{}, layoutapi.ErrUnavailable
	}
	fields, err := current.Fields()
	if err != nil {
		return presentation.FieldCatalog{}, layoutapi.ErrUnavailable
	}
	legacy := make([]presentation.Field, 0, len(fields))
	removed := false
	for _, field := range fields {
		id, fieldErr := field.ID()
		if fieldErr != nil {
			return presentation.FieldCatalog{}, layoutapi.ErrUnavailable
		}
		if id != organizationOwnerIDs {
			legacy = append(legacy, field)
			continue
		}
		if removed || !installationOwnerField(field) {
			return presentation.FieldCatalog{}, layoutapi.ErrUnavailable
		}
		removed = true
	}
	if !removed {
		return presentation.FieldCatalog{}, layoutapi.ErrUnavailable
	}
	candidate, err := presentation.NewFieldCatalog(legacy)
	if err != nil {
		return presentation.FieldCatalog{}, layoutapi.ErrUnavailable
	}
	candidateJSON, candidateSHA, err := catalogEvidence(candidate)
	if err != nil || !sameBytes(published.CatalogJSON, candidateJSON) ||
		!sameDigest(published.CatalogSHA256, candidateSHA) {
		return presentation.FieldCatalog{}, layoutapi.ErrUnavailable
	}
	return candidate, nil
}

func installationOwnerField(field presentation.Field) bool {
	typeID, typeErr := field.TypeID()
	valueProfile, profileErr := field.ValueProfile()
	mode, modeErr := field.Mode()
	return typeErr == nil && profileErr == nil && modeErr == nil &&
		typeID == fieldregistry.TypeUUIDList &&
		valueProfile == fieldregistry.ValueUUIDListV1 &&
		mode == presentation.FieldReadOnly
}
