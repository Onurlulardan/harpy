package portaluser

import (
	"bytes"
	"errors"
	"time"

	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
	"github.com/google/uuid"
)

var (
	ErrInvalid     = errors.New("IAM portal user input is invalid")
	ErrForbidden   = errors.New("IAM portal user operation is forbidden")
	ErrNotFound    = errors.New("IAM portal user was not found")
	ErrConflict    = errors.New("IAM portal user operation conflicts")
	ErrUnavailable = errors.New("IAM portal user operation is unavailable")
)

type Evidence struct {
	OperationID, ActorID, PermissionID uuid.UUID
	RequestSHA256                      [32]byte
	ExpectedTokenVersion               int64
	ExpectedAuthorizationVersion       int64
}

func (value Evidence) Valid() bool {
	return value.OperationID != uuid.Nil && value.ActorID != uuid.Nil &&
		value.PermissionID != uuid.Nil && value.RequestSHA256 != [32]byte{} &&
		value.ExpectedTokenVersion > 0 && value.ExpectedAuthorizationVersion > 0
}

type Identifier struct {
	ID, UserID                       uuid.UUID
	Display, Normalized, LookupKeyID string
	NormalizedSHA256, LookupHMAC     [32]byte
}

func (value Identifier) Valid() bool {
	return value.ID != uuid.Nil && value.Display != "" && value.Normalized != "" &&
		value.LookupKeyID != "" && value.NormalizedSHA256 != [32]byte{} &&
		value.LookupHMAC != [32]byte{}
}

type Subject struct {
	BindingID, AccessModeID, FormID, FormVersionID uuid.UUID
	RecordID, IdentifierFieldID                    uuid.UUID
	Type                                           string
	ContractSHA256                                 [32]byte
}

func (value Subject) Valid() bool {
	return value.BindingID != uuid.Nil && value.AccessModeID != uuid.Nil &&
		value.FormID != uuid.Nil && value.FormVersionID != uuid.Nil &&
		value.RecordID != uuid.Nil && value.IdentifierFieldID != uuid.Nil &&
		value.Type != "" && value.ContractSHA256 != [32]byte{}
}

type CreateCommand struct {
	Evidence          Evidence
	UserID            uuid.UUID
	Profile           identitydomain.ProfileDocument
	Portal, Email     Identifier
	Subject           Subject
	SetupState        string
	RoleNames         []string
	RoleAssignmentIDs []uuid.UUID
}

type RevisionCommand struct {
	Evidence         Evidence
	UserID           uuid.UUID
	ExpectedRevision int64
	Status           string
	AccessModeID     uuid.UUID
	AssignmentID     uuid.UUID
}

type MutationResult struct {
	ID, EmailIdentifierID, PortalIdentifierID uuid.UUID
	Revision, EmailRevision, PortalRevision   int64
	AuthorizationClock                        int64
	Replayed                                  bool
}

func (value MutationResult) Valid() bool {
	return value.ID != uuid.Nil && value.Revision > 0 && value.AuthorizationClock > 0
}

type View struct {
	ID, BindingID, SubjectRecordID uuid.UUID
	Email, Status, AccountType     string
	SubjectType                    string
	Revision                       int64
	SetupRequired                  bool
	CreatedAt, UpdatedAt           time.Time
}

type SubjectLink struct {
	UserID, SubjectRecordID uuid.UUID
	Status, AccountType     string
	Revision                int64
}

type ProjectionAccess struct {
	PrincipalID, TenantID              uuid.UUID
	TokenVersion, AuthorizationVersion int64
	AuthorizationClock                 int64
}

func (value ProjectionAccess) Valid() bool {
	return value.PrincipalID != uuid.Nil && value.TenantID != uuid.Nil &&
		value.TokenVersion > 0 && value.AuthorizationVersion > 0 &&
		value.AuthorizationClock > 0
}

func (value SubjectLink) Valid() bool {
	return value.UserID != uuid.Nil && value.SubjectRecordID != uuid.Nil &&
		(value.Status == "ACTIVE" || value.Status == "PENDING" ||
			value.Status == "INACTIVE") &&
		(value.AccountType == "portal" || value.AccountType == "operator_portal") &&
		value.Revision > 0
}

func (value View) Valid() bool {
	return value.ID != uuid.Nil && value.BindingID != uuid.Nil && value.Email != "" &&
		(value.Status == "ACTIVE" || value.Status == "PENDING" || value.Status == "INACTIVE") &&
		(value.AccountType == "portal" || value.AccountType == "operator_portal") &&
		value.SubjectType != "" && value.SubjectRecordID != uuid.Nil && value.Revision > 0 &&
		!value.CreatedAt.IsZero() && !value.UpdatedAt.Before(value.CreatedAt)
}

func SameDocument(left, right []byte) bool { return bytes.Equal(left, right) }
