package recordkernel

import (
	"bytes"
	"encoding/json"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/google/uuid"
)

type definitionWireV1 struct {
	ContractVersion      string          `json:"contract_version"`
	FormVersionID        string          `json:"form_version_id"`
	SchemaContract       string          `json:"schema_contract"`
	Schema               json.RawMessage `json:"schema"`
	PresentationContract string          `json:"presentation_contract"`
	Presentation         json.RawMessage `json:"presentation"`
}

type definitionWireV2 struct {
	ContractVersion      string          `json:"contract_version"`
	FormVersionID        string          `json:"form_version_id"`
	SchemaContract       string          `json:"schema_contract"`
	Schema               json.RawMessage `json:"schema"`
	RecordContract       string          `json:"record_contract"`
	Record               json.RawMessage `json:"record"`
	PresentationContract string          `json:"presentation_contract"`
	Presentation         json.RawMessage `json:"presentation"`
}

type definitionEnvelope struct {
	contract       string
	versionID      string
	schema         []byte
	recordContract string
	record         []byte
	presentation   []byte
	canonical      []byte
}

func decodeDefinition(document []byte) (definitionEnvelope, error) {
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, MaximumCatalogBytes)
	if err != nil || !bytes.Equal(document, canonical) {
		return definitionEnvelope{}, ErrDefinitionInvalid
	}
	var discriminator struct {
		Contract string `json:"contract_version"`
	}
	if json.Unmarshal(canonical, &discriminator) != nil {
		return definitionEnvelope{}, ErrDefinitionInvalid
	}
	switch discriminator.Contract {
	case DefinitionV1Contract:
		var wire definitionWireV1
		if strictDecode(canonical, &wire) != nil || !validDefinitionV1(wire) {
			return definitionEnvelope{}, ErrDefinitionInvalid
		}
		return definitionEnvelope{wire.ContractVersion, wire.FormVersionID,
			cloneBytes(wire.Schema), "", nil, cloneBytes(wire.Presentation), cloneBytes(canonical)}, nil
	case DefinitionV2Contract:
		var wire definitionWireV2
		if strictDecode(canonical, &wire) != nil || !validDefinitionV2(wire) {
			return definitionEnvelope{}, ErrDefinitionInvalid
		}
		return definitionEnvelope{wire.ContractVersion, wire.FormVersionID,
			cloneBytes(wire.Schema), wire.RecordContract, cloneBytes(wire.Record),
			cloneBytes(wire.Presentation), cloneBytes(canonical)}, nil
	default:
		return definitionEnvelope{}, ErrDefinitionInvalid
	}
}

func validDefinitionV1(wire definitionWireV1) bool {
	return wire.ContractVersion == DefinitionV1Contract && validVersionID(wire.FormVersionID) &&
		wire.SchemaContract == StorageSchemaContractVersion && validObject(wire.Schema) &&
		wire.PresentationContract == PresentationContractVersion && validObject(wire.Presentation)
}

func validDefinitionV2(wire definitionWireV2) bool {
	return wire.ContractVersion == DefinitionV2Contract && validVersionID(wire.FormVersionID) &&
		wire.SchemaContract == StorageSchemaContractVersion && validObject(wire.Schema) &&
		wire.RecordContract == RecordRulesContract && validObject(wire.Record) &&
		wire.PresentationContract == PresentationContractVersion && validObject(wire.Presentation)
}

func validVersionID(raw string) bool {
	value, err := uuid.Parse(raw)
	return err == nil && stableUUIDv7(value) && value.String() == raw
}

func validObject(raw []byte) bool {
	canonical, err := canonicaljsonv1.CanonicalizeObject(raw, MaximumCatalogBytes)
	return err == nil && bytes.Equal(raw, canonical)
}

func cloneBytes(value []byte) []byte { return append([]byte(nil), value...) }
