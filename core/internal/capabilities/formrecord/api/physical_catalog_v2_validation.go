package recordkernel

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"io"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	schemav2 "github.com/Onurlulardan/harpy/contracts/schema/v2"
	"github.com/google/uuid"
)

func compiledPhysicalDocumentV2Valid(
	document []byte, digest [32]byte, catalog Catalog, mappingDigest [32]byte,
) bool {
	if len(document) < 2 || len(document) > maximumPhysicalCatalogBytes ||
		digest == [32]byte{} || sha256.Sum256(document) != digest ||
		catalog.SHA256() == [32]byte{} || mappingDigest == [32]byte{} {
		return false
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, maximumPhysicalCatalogBytes)
	if err != nil || !bytes.Equal(canonical, document) {
		return false
	}
	var decoded physicalCatalogDocumentV2
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	if decoder.Decode(&decoded) != nil {
		return false
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF ||
		!physicalDocumentV2RootMatches(decoded, catalog, mappingDigest) {
		return false
	}
	for index, raw := range decoded.Handles {
		if index > 0 && decoded.Handles[index-1].PhysicalName >= raw.PhysicalName ||
			!compiledPhysicalHandleV2Valid(raw, catalog) {
			return false
		}
	}
	return physicalCoverageV2Valid(catalog, decoded.Handles)
}

func physicalDocumentV2RootMatches(
	document physicalCatalogDocumentV2, catalog Catalog, mappingDigest [32]byte,
) bool {
	recordSHA, recordErr := parseHexDigest(document.RecordCatalogSHA256)
	mappingSHA, mappingErr := parseHexDigest(document.MappingCatalogSHA256)
	return document.Contract == PhysicalCatalogV2Contract &&
		document.PublicationID == catalog.PublicationID().String() &&
		document.PublicationRevision == catalog.PublicationRevision() &&
		document.FormID == catalog.FormID().String() &&
		document.FormVersionID == catalog.FormVersionID().String() &&
		recordErr == nil && recordSHA == catalog.SHA256() && mappingErr == nil &&
		mappingSHA == mappingDigest && len(document.Handles) > 0
}

func compiledPhysicalHandleV2Valid(raw physicalHandleDocumentV2, catalog Catalog) bool {
	base := physicalHandleDocument{LogicalID: raw.LogicalID, FieldID: raw.FieldID,
		Kind: raw.Kind, PhysicalName: raw.PhysicalName, StorageType: raw.StorageType,
		Variant: raw.Variant, Generation: raw.Generation, Profile: raw.Profile}
	if !compiledPhysicalHandleValid(base) {
		return false
	}
	if raw.Kind != physicalRelationTable {
		return physicalTargetBindingAbsent(raw)
	}
	if raw.FieldID == nil || raw.TargetPublicationID == nil || raw.TargetFormID == nil ||
		raw.TargetFormVersionID == nil || raw.TargetPhysicalName == nil {
		return false
	}
	fieldID, err := parsePhysicalDocumentID(*raw.FieldID)
	targetID, targetErr := parsePhysicalDocumentID(*raw.TargetFormID)
	targetPublicationID, publicationErr := parsePhysicalDocumentID(*raw.TargetPublicationID)
	targetVersionID, versionErr := parsePhysicalDocumentID(*raw.TargetFormVersionID)
	field, found := catalog.FieldByID(fieldID)
	expected, deriveErr := schemav2.DeriveIdentifier(schemav2.IdentifierForm, targetID.String())
	return err == nil && targetErr == nil && publicationErr == nil && versionErr == nil &&
		deriveErr == nil && found &&
		field.Relation != nil && field.Relation.TargetFormID == targetID &&
		expected == *raw.TargetPhysicalName && physicalTargetBindingValid(
		raw, catalog, targetPublicationID, targetID, targetVersionID,
	)
}

func physicalTargetBindingAbsent(raw physicalHandleDocumentV2) bool {
	return raw.TargetBindingKind == "" && raw.TargetPublicationID == nil &&
		raw.TargetFormID == nil && raw.TargetFormVersionID == nil &&
		raw.TargetRecordCatalogSHA256 == nil && raw.TargetQueryCatalogSHA256 == nil &&
		raw.TargetPhysicalCatalogSHA256 == nil && raw.TargetPhysicalName == nil
}

func physicalTargetBindingValid(
	raw physicalHandleDocumentV2,
	catalog Catalog,
	publicationID, formID, versionID uuid.UUID,
) bool {
	switch raw.TargetBindingKind {
	case RelationBindingSelf:
		return publicationID == catalog.PublicationID() && formID == catalog.FormID() &&
			versionID == catalog.FormVersionID() && raw.TargetRecordCatalogSHA256 == nil &&
			raw.TargetQueryCatalogSHA256 == nil && raw.TargetPhysicalCatalogSHA256 == nil
	case RelationBindingExternal:
		if formID == catalog.FormID() || raw.TargetRecordCatalogSHA256 == nil ||
			raw.TargetQueryCatalogSHA256 == nil || raw.TargetPhysicalCatalogSHA256 == nil {
			return false
		}
		recordSHA, recordErr := parseHexDigest(*raw.TargetRecordCatalogSHA256)
		querySHA, queryErr := parseHexDigest(*raw.TargetQueryCatalogSHA256)
		physicalSHA, physicalErr := parseHexDigest(*raw.TargetPhysicalCatalogSHA256)
		return recordErr == nil && queryErr == nil && physicalErr == nil &&
			recordSHA != [32]byte{} && querySHA != [32]byte{} && physicalSHA != [32]byte{}
	default:
		return false
	}
}
