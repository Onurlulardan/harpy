package authorization

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"slices"

	"github.com/google/uuid"
)

const maximumPolicyRequirements = 256

type PolicyRequest struct {
	binding       [sha256.Size]byte
	tenantID      uuid.UUID
	actorID       uuid.UUID
	queryIdentity [sha256.Size]byte
	owner         string
	requirements  []Requirement
	validated     bool
}

func newPolicyRequest(
	tenantID uuid.UUID,
	actorID uuid.UUID,
	queryIdentity [sha256.Size]byte,
	owner string,
	requirements []Requirement,
) (PolicyRequest, error) {
	owned := append([]Requirement{}, requirements...)
	slices.SortFunc(owned, compareRequirementIdentity)
	request := PolicyRequest{
		tenantID: tenantID, actorID: actorID, queryIdentity: queryIdentity,
		owner: owner, requirements: owned, validated: true,
	}
	binding, err := request.bindingSHA256()
	if err != nil {
		return PolicyRequest{}, err
	}
	request.binding = binding
	if err := request.Validate(); err != nil {
		return PolicyRequest{}, err
	}
	return request, nil
}

func (request PolicyRequest) Validate() error {
	if !request.validated || !validAuthorizationUUID(request.tenantID) ||
		!validAuthorizationUUID(request.actorID) ||
		request.queryIdentity == [sha256.Size]byte{} || request.owner == "" ||
		len(request.requirements) < 1 ||
		len(request.requirements) > maximumPolicyRequirements ||
		!validPolicyRequirements(request.owner, request.requirements) {
		return fmt.Errorf("REST Query policy request is invalid")
	}
	expected, err := request.bindingSHA256()
	if err != nil || expected == [sha256.Size]byte{} || expected != request.binding {
		return fmt.Errorf("REST Query policy request is invalid")
	}
	return nil
}

func (request PolicyRequest) BindingSHA256() ([sha256.Size]byte, error) {
	if err := request.Validate(); err != nil {
		return [sha256.Size]byte{}, err
	}
	return request.binding, nil
}

func (request PolicyRequest) TenantID() (uuid.UUID, error) {
	if err := request.Validate(); err != nil {
		return uuid.Nil, err
	}
	return request.tenantID, nil
}

func (request PolicyRequest) ActorID() (uuid.UUID, error) {
	if err := request.Validate(); err != nil {
		return uuid.Nil, err
	}
	return request.actorID, nil
}

func (request PolicyRequest) QueryIdentitySHA256() ([sha256.Size]byte, error) {
	if err := request.Validate(); err != nil {
		return [sha256.Size]byte{}, err
	}
	return request.queryIdentity, nil
}

func (request PolicyRequest) Owner() (string, error) {
	if err := request.Validate(); err != nil {
		return "", err
	}
	return request.owner, nil
}

func (request PolicyRequest) Requirements() ([]Requirement, error) {
	if err := request.Validate(); err != nil {
		return nil, err
	}
	return append([]Requirement{}, request.requirements...), nil
}

func (PolicyRequest) String() string   { return authorizationRedaction }
func (PolicyRequest) GoString() string { return authorizationRedaction }
func (PolicyRequest) LogValue() slog.Value {
	return authorizationLogValue()
}
func (request PolicyRequest) Format(state fmt.State, _ rune) {
	writeAuthorizationRedaction(state)
}
func (PolicyRequest) MarshalJSON() ([]byte, error) { return rejectAuthorizationJSON() }

func compareRequirementIdentity(left, right Requirement) int {
	return bytes.Compare(left.id[:], right.id[:])
}
