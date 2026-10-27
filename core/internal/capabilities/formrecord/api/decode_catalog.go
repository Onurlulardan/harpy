package recordkernel

import (
	"bytes"
	"crypto/sha256"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func DecodeCatalog(input DecodeCatalogInput) (Catalog, error) {
	if !input.Binding.valid() || len(input.CanonicalJSON) == 0 ||
		len(input.CanonicalJSON) > MaximumCatalogBytes ||
		input.ExpectedSHA256 == [32]byte{} ||
		sha256.Sum256(input.CanonicalJSON) != input.ExpectedSHA256 {
		return Catalog{}, ErrCatalogInvalid
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		input.CanonicalJSON, MaximumCatalogBytes,
	)
	if err != nil || !bytes.Equal(canonical, input.CanonicalJSON) {
		return Catalog{}, ErrCatalogInvalid
	}
	var document catalogAttestationDocument
	if strictDecode(canonical, &document) != nil {
		return Catalog{}, ErrCatalogInvalid
	}
	catalog, err := decodeCatalogHeader(document, input)
	if err != nil {
		return Catalog{}, err
	}
	fields, byID, byName, err := decodeCatalogFields(
		document.Fields, input.Binding, catalog.limits,
	)
	if err != nil {
		return Catalog{}, err
	}
	catalog.fields, catalog.fieldsByID, catalog.fieldsByName = fields, byID, byName
	catalog.workflowActions, err = compileWorkflowActions(document.WorkflowActions)
	if err != nil {
		return Catalog{}, err
	}
	catalog.defaults, catalog.calculations, catalog.crossFieldValidations, err =
		decodeCatalogRules(document, byID, catalog.limits)
	if err != nil {
		return Catalog{}, err
	}
	catalog.canonical, catalog.sha256 = append([]byte(nil), canonical...), input.ExpectedSHA256
	reencoded, err := catalogDocument(catalog)
	if err != nil || !bytes.Equal(reencoded, canonical) {
		return Catalog{}, ErrCatalogInvalid
	}
	return catalog, nil
}
