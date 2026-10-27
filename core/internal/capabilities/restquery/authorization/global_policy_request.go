package authorization

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"regexp"
	"slices"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

var globalSnapshotOperationPattern = regexp.MustCompile(`^[a-z][a-z0-9_.:-]{0,127}$`)

type GlobalPolicyRequest struct {
	binding       [sha256.Size]byte
	environmentID uuid.UUID
	actorID       uuid.UUID
	queryIdentity [sha256.Size]byte
	owner         string
	operationID   string
	snapshotAt    time.Time
	requirements  []Requirement
	public        bool
	validated     bool
}

func newGlobalPolicyRequest(
	scope transaction.GlobalScope,
	queryIdentity [sha256.Size]byte,
	snapshotAt time.Time,
	owner string,
	requirements []Requirement,
) (GlobalPolicyRequest, error) {
	owned := append([]Requirement{}, requirements...)
	slices.SortFunc(owned, compareRequirementIdentity)
	request := GlobalPolicyRequest{
		environmentID: scope.EnvironmentID(), actorID: scope.ActorID(),
		queryIdentity: queryIdentity, owner: owner, operationID: scope.OperationID(),
		snapshotAt: snapshotAt.UTC().Truncate(time.Second), requirements: owned,
		public: scope.ActorID() == uuid.Nil, validated: true,
	}
	binding, err := request.bindingSHA256()
	request.binding = binding
	if err != nil || request.Validate() != nil {
		return GlobalPolicyRequest{}, errPolicyEvaluationContract
	}
	return request, nil
}

func (request GlobalPolicyRequest) Validate() error {
	if !request.validated || !validAuthorizationUUID(request.environmentID) ||
		request.queryIdentity == [sha256.Size]byte{} || request.owner == "" ||
		!globalSnapshotOperationPattern.MatchString(request.operationID) ||
		request.snapshotAt.IsZero() || request.snapshotAt.Location() != time.UTC ||
		len(request.requirements) < 1 || len(request.requirements) > maximumPolicyRequirements ||
		!validPolicyRequirements(request.owner, request.requirements) ||
		request.public != (request.actorID == uuid.Nil) ||
		!request.public && !validAuthorizationUUID(request.actorID) {
		return fmt.Errorf("REST Query global policy request is invalid")
	}
	expected, err := request.bindingSHA256()
	if err != nil || expected == [sha256.Size]byte{} || expected != request.binding {
		return fmt.Errorf("REST Query global policy request is invalid")
	}
	return nil
}

func (GlobalPolicyRequest) String() string       { return authorizationRedaction }
func (GlobalPolicyRequest) GoString() string     { return authorizationRedaction }
func (GlobalPolicyRequest) LogValue() slog.Value { return authorizationLogValue() }
func (request GlobalPolicyRequest) Format(state fmt.State, _ rune) {
	writeAuthorizationRedaction(state)
}
func (GlobalPolicyRequest) MarshalJSON() ([]byte, error) { return rejectAuthorizationJSON() }
