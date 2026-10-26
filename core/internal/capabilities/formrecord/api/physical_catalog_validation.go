package recordkernel

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"io"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/google/uuid"
)

func compiledPhysicalDocumentValid(
	document []byte, digest [32]byte, catalog Catalog, mappingDigest [32]byte,
) bool {
	if len(document) < 2 || len(document) > maximumPhysicalCatalogBytes ||
		digest == [32]byte{} || sha256.Sum256(document) != digest ||
		catalog.SHA256() == [32]byte{} || mappingDigest == [32]byte{} {
		return false
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		document, maximumPhysicalCatalogBytes,
	)
	if err != nil || !bytes.Equal(canonical, document) {
		return false
	}
	var decoded physicalCatalogDocument
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	if decoder.Decode(&decoded) != nil {
		return false
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF ||
		!physicalDocumentRootMatches(decoded, catalog, mappingDigest) {
		return false
	}
	for index, raw := range decoded.Handles {
		if index > 0 && decoded.Handles[index-1].PhysicalName >= raw.PhysicalName ||
			!compiledPhysicalHandleValid(raw) {
			return false
		}
	}
	return physicalCoverageValid(catalog, catalog.Fields(), decoded.Handles)
}

func physicalDocumentRootMatches(
	document physicalCatalogDocument, catalog Catalog, mappingDigest [32]byte,
) bool {
	recordSHA, recordErr := parseHexDigest(document.RecordCatalogSHA256)
	mappingSHA, mappingErr := parseHexDigest(document.MappingCatalogSHA256)
	return document.Contract == PhysicalCatalogContract &&
		document.PublicationID == catalog.PublicationID().String() &&
		document.PublicationRevision == catalog.PublicationRevision() &&
		document.FormID == catalog.FormID().String() &&
		document.FormVersionID == catalog.FormVersionID().String() &&
		recordErr == nil && recordSHA == catalog.SHA256() && mappingErr == nil &&
		mappingSHA == mappingDigest && len(document.Handles) > 0
}

func compiledPhysicalHandleValid(document physicalHandleDocument) bool {
	logicalID, err := parsePhysicalDocumentID(document.LogicalID)
	if err != nil || document.Generation < 1 {
		return false
	}
	fieldIDPresent := document.FieldID != nil
	if fieldIDPresent {
		if _, err = parsePhysicalDocumentID(*document.FieldID); err != nil {
			return false
		}
	}
	mapping := PhysicalMapping{LogicalID: logicalID, Kind: document.Kind,
		PhysicalName: document.PhysicalName, StorageType: document.StorageType,
		Variant: document.Variant, Generation: document.Generation, Profile: document.Profile}
	return mapping.Valid() &&
		(document.Kind == physicalRelationTable) == fieldIDPresent
}

func parsePhysicalDocumentID(value string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(value)
	if err != nil || parsed.String() != value || !stableUUIDv7(parsed) {
		return uuid.Nil, ErrCatalogInvalid
	}
	return parsed, nil
}
