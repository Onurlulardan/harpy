package recordkernel

import (
	"crypto/sha256"
	"slices"

	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/google/uuid"
)

// CompilePhysicalCatalogDocumentV2 seals every runtime identifier, including
// relation target tables, into publication evidence.
func CompilePhysicalCatalogDocumentV2(
	catalog Catalog,
	mappingSHA [32]byte,
	mappings []PhysicalMapping,
	external []ExternalRelationCatalogBinding,
) ([]byte, [32]byte, error) {
	recomputed, err := DigestMappings(mappings)
	if catalog.SHA256() == [32]byte{} || mappingSHA == [32]byte{} || len(mappings) == 0 ||
		err != nil || recomputed != mappingSHA {
		return nil, [32]byte{}, ErrCatalogInvalid
	}
	fields := catalog.Fields()
	bindings, err := indexExternalRelationBindings(catalog, external)
	if err != nil {
		return nil, [32]byte{}, err
	}
	usedBindings := make(map[uuid.UUID]struct{}, len(bindings))
	byID := make(map[uuid.UUID]Field, len(fields))
	relations := make(map[uuid.UUID]Field)
	for _, field := range fields {
		byID[field.ID] = field
		if field.Relation != nil {
			relations[field.Relation.RelationID] = field
		}
	}
	handles := make([]physicalHandleDocumentV2, 0, len(mappings))
	for _, mapping := range mappings {
		base, included, buildErr := physicalDocumentFor(catalog, mapping, byID, relations)
		if buildErr != nil {
			return nil, [32]byte{}, buildErr
		}
		if !included {
			continue
		}
		handle := physicalHandleV2FromBase(base)
		if mapping.Kind == physicalRelationTable {
			field := relations[mapping.LogicalID]
			targetID := field.Relation.TargetFormID.String()
			targetName, deriveErr := schemav2.DeriveIdentifier(schemav2.IdentifierForm, targetID)
			if deriveErr != nil {
				return nil, [32]byte{}, ErrCatalogInvalid
			}
			handle.TargetFormID, handle.TargetPhysicalName = &targetID, &targetName
			if field.Relation.TargetFormID == catalog.FormID() {
				publicationID, versionID := catalog.PublicationID().String(),
					catalog.FormVersionID().String()
				handle.TargetBindingKind = RelationBindingSelf
				handle.TargetPublicationID = &publicationID
				handle.TargetFormVersionID = &versionID
			} else {
				binding, found := bindings[field.Relation.TargetFormID]
				if !found {
					return nil, [32]byte{}, ErrCatalogInvalid
				}
				applyExternalRelationBinding(&handle, binding)
				usedBindings[binding.FormID] = struct{}{}
			}
		}
		handles = append(handles, handle)
	}
	slices.SortFunc(handles, func(left, right physicalHandleDocumentV2) int {
		return compareText(left.PhysicalName, right.PhysicalName)
	})
	if len(usedBindings) != len(bindings) {
		return nil, [32]byte{}, ErrCatalogInvalid
	}
	document := physicalCatalogDocumentV2{Contract: PhysicalCatalogV2Contract,
		PublicationID:       catalog.PublicationID().String(),
		PublicationRevision: catalog.PublicationRevision(), FormID: catalog.FormID().String(),
		FormVersionID:        catalog.FormVersionID().String(),
		RecordCatalogSHA256:  hexDigest(catalog.SHA256()),
		MappingCatalogSHA256: hexDigest(mappingSHA), Handles: handles}
	canonical, err := canonicalObject(document, maximumPhysicalCatalogBytes)
	if err != nil {
		return nil, [32]byte{}, ErrCatalogInvalid
	}
	digest := sha256.Sum256(canonical)
	if !compiledPhysicalDocumentV2Valid(canonical, digest, catalog, mappingSHA) {
		return nil, [32]byte{}, ErrCatalogInvalid
	}
	return append([]byte(nil), canonical...), digest, nil
}

func indexExternalRelationBindings(
	catalog Catalog,
	values []ExternalRelationCatalogBinding,
) (map[uuid.UUID]ExternalRelationCatalogBinding, error) {
	result := make(map[uuid.UUID]ExternalRelationCatalogBinding, len(values))
	for _, binding := range values {
		if !binding.valid() || binding.FormID == catalog.FormID() {
			return nil, ErrCatalogInvalid
		}
		if _, found := result[binding.FormID]; found {
			return nil, ErrCatalogInvalid
		}
		result[binding.FormID] = binding
	}
	return result, nil
}

func applyExternalRelationBinding(
	handle *physicalHandleDocumentV2,
	binding ExternalRelationCatalogBinding,
) {
	publicationID, versionID := binding.PublicationID.String(), binding.FormVersionID.String()
	recordSHA, querySHA := hexDigest(binding.RecordCatalogSHA256),
		hexDigest(binding.QueryCatalogSHA256)
	physicalSHA := hexDigest(binding.PhysicalCatalogSHA256)
	handle.TargetBindingKind = RelationBindingExternal
	handle.TargetPublicationID, handle.TargetFormVersionID = &publicationID, &versionID
	handle.TargetRecordCatalogSHA256 = &recordSHA
	handle.TargetQueryCatalogSHA256 = &querySHA
	handle.TargetPhysicalCatalogSHA256 = &physicalSHA
}

func physicalHandleV2FromBase(base physicalHandleDocument) physicalHandleDocumentV2 {
	return physicalHandleDocumentV2{LogicalID: base.LogicalID, FieldID: base.FieldID,
		Kind: base.Kind, PhysicalName: base.PhysicalName, StorageType: base.StorageType,
		Variant: base.Variant, Generation: base.Generation, Profile: base.Profile}
}
