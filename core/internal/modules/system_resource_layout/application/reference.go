package application

import (
	"context"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	"github.com/google/uuid"
)

type ReferenceTarget struct {
	Resource layoutapi.ResourceKind
	FieldID  uuid.UUID
}

func (target ReferenceTarget) Validate() error {
	if !target.Resource.Valid() || !stableUUID(target.FieldID) {
		return layoutapi.ErrInvalid
	}
	return nil
}

type ReferenceEvidence struct {
	Target         ReferenceTarget
	ManifestSHA256 [32]byte
	CheckedOwners  int32
	LiveReferences int64
}

func (evidence ReferenceEvidence) Validate() error {
	if evidence.Target.Validate() != nil || evidence.ManifestSHA256 == [32]byte{} ||
		evidence.CheckedOwners != 1 || evidence.LiveReferences < 0 {
		return layoutapi.ErrUnavailable
	}
	return nil
}

type ReferenceStore interface {
	FieldReferenceEvidence(
		context.Context,
		ReferenceTarget,
	) (ReferenceEvidence, error)
}
