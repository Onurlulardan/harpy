package recordkernel

import (
	"github.com/google/uuid"
)

type PublishedField struct {
	FieldID    uuid.UUID
	PublicName string
	Type       string
	Role       string
	Generation uint32
}

type CatalogBinding struct {
	TenantID                uuid.UUID
	PublicationID           uuid.UUID
	PublicationRevision     int64
	FormID                  uuid.UUID
	FormVersionID           uuid.UUID
	SourceDocumentSHA256    [32]byte
	RecordRulesSHA256       [32]byte
	RecordAttestationSHA256 [32]byte
	FieldCatalogSHA256      [32]byte
	BudgetProfileID         string
	BudgetSHA256            [32]byte
	RegistrySHA256          [32]byte
	MappingCatalogSHA256    [32]byte
	Fields                  []PublishedField
}

func (binding CatalogBinding) valid() bool {
	return tenantUUID(binding.TenantID) && stableUUIDv7(binding.PublicationID) &&
		binding.PublicationRevision > 0 && stableUUIDv7(binding.FormID) &&
		stableUUIDv7(binding.FormVersionID) && binding.FieldCatalogSHA256 != [32]byte{} &&
		binding.SourceDocumentSHA256 != [32]byte{} &&
		binding.RecordRulesSHA256 != [32]byte{} &&
		binding.RecordAttestationSHA256 != [32]byte{} &&
		binding.BudgetProfileID != "" && binding.BudgetSHA256 != [32]byte{} &&
		binding.RegistrySHA256 != [32]byte{} &&
		binding.MappingCatalogSHA256 != [32]byte{} && len(binding.Fields) > 0
}

func tenantUUID(value uuid.UUID) bool {
	return value != uuid.Nil && value.Variant() == uuid.RFC4122
}

func stableUUIDv7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == 7 && value.Variant() == uuid.RFC4122
}
