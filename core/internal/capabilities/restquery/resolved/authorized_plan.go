package resolved

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"slices"
	"time"

	"github.com/google/uuid"
)

const AuthorizationPolicySchemaV1 = "restquery_authorization_v1"

type AuthorizedPlan struct {
	spec                 QuerySpec
	snapshot             SnapshotBinding
	authorizationContext [sha256.Size]byte
	visibilities         []ResourceVisibility
	mandatoryCanonical   []byte
	stateSeal            [sha256.Size]byte
	validated            bool
}

func MintAuthorizedPlan(
	spec QuerySpec,
	tenantID uuid.UUID,
	actorID uuid.UUID,
	authorizationContext [sha256.Size]byte,
	visibilities []ResourceVisibility,
) (AuthorizedPlan, error) {
	snapshot, err := newTenantSnapshotBinding(tenantID, actorID)
	if err != nil {
		return AuthorizedPlan{}, err
	}
	return mintAuthorizedPlan(spec, snapshot, authorizationContext, visibilities)
}

func MintGlobalAuthorizedPlan(
	spec QuerySpec,
	environmentID uuid.UUID,
	actorID uuid.UUID,
	operationID string,
	public bool,
	snapshotAt time.Time,
	authorizationContext [sha256.Size]byte,
	visibilities []ResourceVisibility,
) (AuthorizedPlan, error) {
	snapshot, err := newGlobalSnapshotBinding(
		environmentID, actorID, operationID, public, snapshotAt,
	)
	if err != nil {
		return AuthorizedPlan{}, err
	}
	return mintAuthorizedPlan(spec, snapshot, authorizationContext, visibilities)
}

func mintAuthorizedPlan(
	spec QuerySpec,
	snapshot SnapshotBinding,
	authorizationContext [sha256.Size]byte,
	visibilities []ResourceVisibility,
) (AuthorizedPlan, error) {
	owned := cloneResourceVisibilities(visibilities)
	slices.SortFunc(owned, func(left, right ResourceVisibility) int {
		return bytes.Compare(left.requirement[:], right.requirement[:])
	})
	mandatory, err := canonicalMandatoryVisibilities(owned)
	if err != nil {
		return AuthorizedPlan{}, err
	}
	plan := AuthorizedPlan{
		spec: spec, snapshot: snapshot,
		authorizationContext: authorizationContext,
		visibilities:         owned, mandatoryCanonical: mandatory,
		validated: true,
	}
	plan.stateSeal = authorizedPlanSeal(plan)
	if err := plan.Validate(); err != nil {
		return AuthorizedPlan{}, err
	}
	return plan, nil
}

func (plan AuthorizedPlan) Validate() error {
	if !plan.validated || plan.spec.Validate() != nil ||
		plan.snapshot.Validate() != nil ||
		plan.authorizationContext == [sha256.Size]byte{} ||
		len(plan.visibilities) < 1 || len(plan.visibilities) > 4 ||
		len(plan.mandatoryCanonical) < 1 ||
		len(plan.mandatoryCanonical) > maximumMandatoryCanonicalBytes ||
		!visibilitiesMatchQuery(plan.spec, plan.visibilities) {
		return fmt.Errorf("resolved REST Query authorized plan is invalid")
	}
	canonical, err := canonicalMandatoryVisibilities(plan.visibilities)
	if err != nil || !bytes.Equal(canonical, plan.mandatoryCanonical) ||
		plan.stateSeal == [sha256.Size]byte{} ||
		plan.stateSeal != authorizedPlanSeal(plan) {
		return fmt.Errorf("resolved REST Query authorized plan is invalid")
	}
	return nil
}

func validPlanUUID(value uuid.UUID) bool {
	return value != uuid.Nil && value.Variant() == uuid.RFC4122
}

func (AuthorizedPlan) String() string   { return valueRedaction }
func (AuthorizedPlan) GoString() string { return valueRedaction }
func (plan AuthorizedPlan) Format(state fmt.State, _ rune) {
	writeRedaction(state)
}
func (AuthorizedPlan) MarshalJSON() ([]byte, error) {
	return rejectDirectJSONMarshal()
}
