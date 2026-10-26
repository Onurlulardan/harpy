package recordkernel

import (
	"crypto/sha256"
	"slices"

	"github.com/google/uuid"
)

func CompilePhysicalCatalogDocument(
	catalog Catalog, mappingSHA [32]byte, mappings []PhysicalMapping,
) ([]byte, [32]byte, error) {
	recomputed, err := DigestMappings(mappings)
	if catalog.SHA256() == [32]byte{} || mappingSHA == [32]byte{} || len(mappings) == 0 ||
		err != nil || recomputed != mappingSHA {
		return nil, [32]byte{}, ErrCatalogInvalid
	}
	fields := catalog.Fields()
	byID := make(map[uuid.UUID]Field, len(fields))
	relations := make(map[uuid.UUID]Field)
	for _, field := range fields {
		byID[field.ID] = field
		if field.Relation != nil {
			relations[field.Relation.RelationID] = field
		}
	}
	documents := make([]physicalHandleDocument, 0, len(mappings))
	for _, mapping := range mappings {
		document, included, err := physicalDocumentFor(catalog, mapping, byID, relations)
		if err != nil {
			return nil, [32]byte{}, err
		}
		if included {
			documents = append(documents, document)
		}
	}
	slices.SortFunc(documents, func(left, right physicalHandleDocument) int {
		return compareText(left.PhysicalName, right.PhysicalName)
	})
	if !physicalCoverageValid(catalog, fields, documents) {
		return nil, [32]byte{}, ErrCatalogInvalid
	}
	document := physicalCatalogDocument{
		Contract: PhysicalCatalogContract, PublicationID: catalog.PublicationID().String(),
		PublicationRevision: catalog.PublicationRevision(), FormID: catalog.FormID().String(),
		FormVersionID:        catalog.FormVersionID().String(),
		RecordCatalogSHA256:  hexDigest(catalog.SHA256()),
		MappingCatalogSHA256: hexDigest(mappingSHA), Handles: documents,
	}
	canonical, err := canonicalObject(document, maximumPhysicalCatalogBytes)
	if err != nil {
		return nil, [32]byte{}, ErrCatalogInvalid
	}
	digest := sha256.Sum256(canonical)
	if !compiledPhysicalDocumentValid(canonical, digest, catalog, mappingSHA) {
		return nil, [32]byte{}, ErrCatalogInvalid
	}
	return append([]byte(nil), canonical...), digest, nil
}

func physicalDocumentFor(
	catalog Catalog, mapping PhysicalMapping, fields map[uuid.UUID]Field,
	relations map[uuid.UUID]Field,
) (physicalHandleDocument, bool, error) {
	if !mapping.Valid() {
		return physicalHandleDocument{}, false, ErrCatalogInvalid
	}
	document := physicalHandleDocument{LogicalID: mapping.LogicalID.String(),
		Kind: mapping.Kind, PhysicalName: mapping.PhysicalName,
		StorageType: mapping.StorageType, Variant: mapping.Variant,
		Generation: mapping.Generation, Profile: mapping.Profile}
	switch mapping.Kind {
	case physicalFormTable:
		if mapping.LogicalID != catalog.FormID() {
			return physicalHandleDocument{}, false, ErrCatalogInvalid
		}
	case physicalFieldColumn:
		field, found := fields[mapping.LogicalID]
		if !found || field.Role == RoleRelation || field.Generation != mapping.Generation {
			return physicalHandleDocument{}, false, ErrCatalogInvalid
		}
	case physicalRelationTable:
		field, found := relations[mapping.LogicalID]
		if !found || field.Relation == nil || field.Generation != mapping.Generation {
			return physicalHandleDocument{}, false, ErrCatalogInvalid
		}
		fieldID := field.ID.String()
		document.FieldID = &fieldID
	case "index", "constraint":
		return physicalHandleDocument{}, false, nil
	default:
		return physicalHandleDocument{}, false, ErrCatalogInvalid
	}
	return document, true, nil
}
