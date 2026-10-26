package maintenance

import (
	"crypto/sha256"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

type profilePage struct {
	planID         uuid.UUID
	resource       profileapp.ResourceKind
	sourceSchemaID uuid.UUID
	targetSchemaID uuid.UUID
	pageNumber     int
	afterID        *uuid.UUID
	throughID      uuid.UUID
	expectedCount  int
	sourceSHA      []byte
	targetSHA      []byte
}

func (page profilePage) valid() bool {
	return page.planID != uuid.Nil && page.resource.Valid() &&
		page.sourceSchemaID != uuid.Nil && page.targetSchemaID != uuid.Nil &&
		page.sourceSchemaID != page.targetSchemaID && page.pageNumber > 0 &&
		page.throughID != uuid.Nil && page.expectedCount >= 0 &&
		page.expectedCount <= 250 && len(page.sourceSHA) == sha256.Size &&
		len(page.targetSHA) == sha256.Size &&
		(page.afterID == nil || page.afterID.String() < page.throughID.String())
}

type profileMapping struct {
	sourceID    uuid.UUID
	targetID    uuid.UUID
	kind        string
	semanticSHA []byte
}

func (mapping profileMapping) valid() bool {
	return mapping.sourceID != uuid.Nil && mapping.targetID != uuid.Nil &&
		(mapping.kind == "copy_v1" || mapping.kind == "copy_normalize_v1") &&
		len(mapping.semanticSHA) == sha256.Size
}

type profileResource struct {
	id             uuid.UUID
	actualSchemaID uuid.UUID
	revision       int64
	document       []byte
	sha            []byte
	preconverted   bool
	plannedTarget  []byte
}

func (resource *profileResource) wipe() {
	clear(resource.document)
	clear(resource.sha)
	clear(resource.plannedTarget)
}
