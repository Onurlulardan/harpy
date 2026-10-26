package resolved

import (
	"fmt"
	"log/slog"
	"regexp"
	"time"

	"github.com/google/uuid"
)

type SnapshotScope uint8

const (
	SnapshotScopeTenant SnapshotScope = iota + 1
	SnapshotScopeGlobal
)

type SnapshotAudience uint8

const (
	SnapshotAudienceTenantActor SnapshotAudience = iota + 1
	SnapshotAudienceGlobalActor
	SnapshotAudienceGlobalPublic
)

var snapshotOperationPattern = regexp.MustCompile(`^[a-z][a-z0-9_.-]{0,127}$`)

type SnapshotBinding struct {
	scope       SnapshotScope
	scopeID     uuid.UUID
	subjectID   uuid.UUID
	audience    SnapshotAudience
	operationID string
	snapshotAt  time.Time
	validated   bool
}

func newTenantSnapshotBinding(tenantID, actorID uuid.UUID) (SnapshotBinding, error) {
	return newSnapshotBinding(SnapshotBinding{
		scope: SnapshotScopeTenant, scopeID: tenantID, subjectID: actorID,
		audience: SnapshotAudienceTenantActor, operationID: "restquery.execute",
	})
}

// NewTenantSnapshotBinding constructs non-authorizing tenant query provenance.
// Only MintAuthorizedPlan turns this identity into an executable plan.
func NewTenantSnapshotBinding(tenantID, actorID uuid.UUID) (SnapshotBinding, error) {
	return newTenantSnapshotBinding(tenantID, actorID)
}

func newGlobalSnapshotBinding(
	environmentID, actorID uuid.UUID,
	operationID string,
	public bool,
	snapshotAt time.Time,
) (SnapshotBinding, error) {
	audience := SnapshotAudienceGlobalActor
	if public {
		audience = SnapshotAudienceGlobalPublic
	}
	return newSnapshotBinding(SnapshotBinding{
		scope: SnapshotScopeGlobal, scopeID: environmentID, subjectID: actorID,
		audience: audience, operationID: operationID,
		snapshotAt: snapshotAt.UTC().Truncate(time.Second),
	})
}

func newSnapshotBinding(binding SnapshotBinding) (SnapshotBinding, error) {
	binding.validated = true
	if binding.Validate() != nil {
		return SnapshotBinding{}, fmt.Errorf("REST Query snapshot binding is invalid")
	}
	return binding, nil
}

func (binding SnapshotBinding) Validate() error {
	if !binding.validated || !validPlanUUID(binding.scopeID) ||
		!snapshotOperationPattern.MatchString(binding.operationID) {
		return fmt.Errorf("REST Query snapshot binding is invalid")
	}
	switch binding.scope {
	case SnapshotScopeTenant:
		if binding.audience != SnapshotAudienceTenantActor ||
			!validPlanUUID(binding.subjectID) || !binding.snapshotAt.IsZero() {
			return fmt.Errorf("REST Query snapshot binding is invalid")
		}
	case SnapshotScopeGlobal:
		if binding.snapshotAt.IsZero() || binding.snapshotAt.Location() != time.UTC {
			return fmt.Errorf("REST Query snapshot binding is invalid")
		}
		if binding.audience == SnapshotAudienceGlobalActor {
			if !validPlanUUID(binding.subjectID) {
				return fmt.Errorf("REST Query snapshot binding is invalid")
			}
		} else if binding.audience != SnapshotAudienceGlobalPublic || binding.subjectID != uuid.Nil {
			return fmt.Errorf("REST Query snapshot binding is invalid")
		}
	default:
		return fmt.Errorf("REST Query snapshot binding is invalid")
	}
	return nil
}

func (SnapshotBinding) String() string                         { return valueRedaction }
func (SnapshotBinding) GoString() string                       { return valueRedaction }
func (SnapshotBinding) LogValue() slog.Value                   { return slog.StringValue(valueRedaction) }
func (binding SnapshotBinding) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (SnapshotBinding) MarshalJSON() ([]byte, error)           { return rejectDirectJSONMarshal() }
