package api

import (
	"context"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	layoutmodule "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type layoutReferencePort interface {
	FieldReferenceEvidence(
		context.Context,
		platformtx.TenantScope,
		layoutmodule.ReferenceTarget,
	) (layoutmodule.ReferenceEvidence, error)
}

type profileReferences struct {
	layout layoutReferencePort
}

func newProfileReferences(layout layoutReferencePort) iammodule.ProfileReferencePort {
	return &profileReferences{layout: layout}
}

func (references *profileReferences) AssertRetirable(
	ctx context.Context,
	scope platformtx.TenantScope,
	target iammodule.ProfileRetirementTarget,
) (iammodule.ProfileReferenceEvidence, error) {
	resource, err := layoutResource(target.Resource)
	if err != nil {
		return iammodule.ProfileReferenceEvidence{}, err
	}
	evidence, err := references.layout.FieldReferenceEvidence(
		ctx, scope, layoutmodule.ReferenceTarget{
			Resource: resource,
			FieldID:  target.FieldID,
		},
	)
	if err != nil {
		return iammodule.ProfileReferenceEvidence{}, err
	}
	return iammodule.ProfileReferenceEvidence{
		Target: target, ManifestSHA256: evidence.ManifestSHA256,
		CheckedOwners:  evidence.CheckedOwners,
		LiveReferences: evidence.LiveReferences,
	}, nil
}

func layoutResource(
	resource iammodule.ProfileResourceKind,
) (layoutmodule.ResourceKind, error) {
	switch resource {
	case iammodule.ProfileResourceUser:
		return layoutmodule.ResourceUser, nil
	case iammodule.ProfileResourceOrganization:
		return layoutmodule.ResourceOrganization, nil
	case iammodule.ProfileResourceRole:
		return layoutmodule.ResourceRole, nil
	default:
		return "", layoutmodule.ErrInvalid
	}
}
