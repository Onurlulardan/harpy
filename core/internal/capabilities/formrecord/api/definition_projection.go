package recordkernel

import (
	"encoding/json"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func (value definitionEnvelope) v1Projection() ([]byte, error) {
	wire := definitionWireV1{
		ContractVersion:      DefinitionV1Contract,
		FormVersionID:        value.versionID,
		SchemaContract:       StorageSchemaContractVersion,
		Schema:               cloneBytes(value.schema),
		PresentationContract: PresentationContractVersion,
		Presentation:         cloneBytes(value.presentation),
	}
	encoded, err := json.Marshal(wire)
	if err != nil {
		return nil, ErrDefinitionInvalid
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(encoded, MaximumCatalogBytes)
	if err != nil {
		return nil, ErrDefinitionInvalid
	}
	return canonical, nil
}
