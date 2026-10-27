package v1adoption

import "github.com/google/uuid"

const MaximumPageSize = 500

type Plan struct {
	RunID                uuid.UUID
	TenantID             uuid.UUID
	SourceGeneration     string
	SourceManifestDigest [32]byte
	PageSize             int
}

func (value Plan) Validate() error {
	if value.RunID == uuid.Nil || value.TenantID == uuid.Nil ||
		value.SourceGeneration == "" || len(value.SourceGeneration) > 128 ||
		value.SourceManifestDigest == [32]byte{} || value.PageSize < 1 ||
		value.PageSize > MaximumPageSize {
		return ErrInvalid
	}
	return nil
}
