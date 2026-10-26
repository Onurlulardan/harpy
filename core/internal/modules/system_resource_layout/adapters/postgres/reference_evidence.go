package postgres

import (
	"context"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	layoutapp "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/application"
)

func (store *layoutStore) FieldReferenceEvidence(
	ctx context.Context,
	target layoutapp.ReferenceTarget,
) (layoutapp.ReferenceEvidence, error) {
	if !store.queryUsable(ctx) || target.Validate() != nil {
		return layoutapp.ReferenceEvidence{}, layoutapi.ErrInvalid
	}
	var live int64
	var digest []byte
	var owners int32
	err := store.queryRow(
		ctx, referenceEvidenceSQL, string(target.Resource), target.FieldID,
	).Scan(&live, &digest, &owners)
	if err != nil {
		return layoutapp.ReferenceEvidence{}, mapLayoutError(err)
	}
	if len(digest) != 32 {
		return layoutapp.ReferenceEvidence{}, layoutapi.ErrUnavailable
	}
	evidence := layoutapp.ReferenceEvidence{
		Target: target, CheckedOwners: owners, LiveReferences: live,
	}
	copy(evidence.ManifestSHA256[:], digest)
	if evidence.Validate() != nil {
		return layoutapp.ReferenceEvidence{}, layoutapi.ErrUnavailable
	}
	return evidence, nil
}

var _ layoutapp.ReferenceStore = (*layoutStore)(nil)
