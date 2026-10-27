package resolved

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"
)

func (plan AuthorizedPlan) Spec() (QuerySpec, error) {
	if err := plan.Validate(); err != nil {
		return QuerySpec{}, err
	}
	return plan.spec, nil
}

func (plan AuthorizedPlan) TenantID() (uuid.UUID, error) {
	if err := plan.Validate(); err != nil || plan.snapshot.scope != SnapshotScopeTenant {
		return uuid.Nil, fmt.Errorf("resolved REST Query plan is not tenant-bound")
	}
	return plan.snapshot.scopeID, nil
}

func (plan AuthorizedPlan) ActorID() (uuid.UUID, error) {
	if err := plan.Validate(); err != nil || plan.snapshot.subjectID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("resolved REST Query plan has no actor")
	}
	return plan.snapshot.subjectID, nil
}

func (plan AuthorizedPlan) SnapshotBinding() (SnapshotBinding, error) {
	if err := plan.Validate(); err != nil {
		return SnapshotBinding{}, err
	}
	return plan.snapshot, nil
}

func (plan AuthorizedPlan) AuthorizationContextSHA256() (
	[sha256.Size]byte,
	error,
) {
	if err := plan.Validate(); err != nil {
		return [sha256.Size]byte{}, err
	}
	return plan.authorizationContext, nil
}

func (plan AuthorizedPlan) Visibilities() ([]ResourceVisibility, error) {
	if err := plan.Validate(); err != nil {
		return nil, err
	}
	return cloneResourceVisibilities(plan.visibilities), nil
}

func (plan AuthorizedPlan) MandatoryCanonical() ([]byte, error) {
	if err := plan.Validate(); err != nil {
		return nil, err
	}
	return bytes.Clone(plan.mandatoryCanonical), nil
}

func (plan AuthorizedPlan) PolicySchema() (string, error) {
	if err := plan.Validate(); err != nil {
		return "", err
	}
	return AuthorizationPolicySchemaV1, nil
}
