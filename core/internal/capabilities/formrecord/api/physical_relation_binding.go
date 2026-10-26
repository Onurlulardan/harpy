package recordkernel

import "github.com/google/uuid"

const (
	RelationBindingSelf     = "self"
	RelationBindingExternal = "external"
)

type ExternalRelationCatalogBinding struct {
	PublicationID         uuid.UUID
	FormID                uuid.UUID
	FormVersionID         uuid.UUID
	RecordCatalogSHA256   [32]byte
	QueryCatalogSHA256    [32]byte
	PhysicalCatalogSHA256 [32]byte
}

func (binding ExternalRelationCatalogBinding) valid() bool {
	return stablePhysicalBindingID(binding.PublicationID) &&
		stablePhysicalBindingID(binding.FormID) &&
		stablePhysicalBindingID(binding.FormVersionID) &&
		binding.RecordCatalogSHA256 != [32]byte{} &&
		binding.QueryCatalogSHA256 != [32]byte{} &&
		binding.PhysicalCatalogSHA256 != [32]byte{}
}

func stablePhysicalBindingID(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == 7 && value.Variant() == uuid.RFC4122
}
