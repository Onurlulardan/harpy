package profile

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type SelfEvidence struct {
	ActorID                      uuid.UUID
	ExpectedTokenVersion         int64
	ExpectedAuthorizationVersion int64
}

func (evidence SelfEvidence) Validate() error {
	if !stableID(evidence.ActorID) || evidence.ExpectedTokenVersion < 1 ||
		evidence.ExpectedAuthorizationVersion < 1 {
		return ErrInvalidCommand
	}
	return nil
}

type SelfRole struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type SelfOrganization struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}

type SelfRecord struct {
	ID            uuid.UUID
	Email         string
	Status        string
	Revision      int64
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Roles         []SelfRole
	Organizations []SelfOrganization
}

func (record SelfRecord) Validate(actorID uuid.UUID) error {
	if !stableID(actorID) || record.ID != actorID || record.Email == "" ||
		record.Status != "active" || record.Revision < 1 ||
		record.CreatedAt.IsZero() || record.UpdatedAt.Before(record.CreatedAt) ||
		record.Roles == nil || record.Organizations == nil {
		return ErrEvidenceMissing
	}
	for _, role := range record.Roles {
		if !stableID(role.ID) || role.Name == "" {
			return ErrEvidenceMissing
		}
	}
	for _, organization := range record.Organizations {
		if !stableID(organization.ID) || organization.Name == "" ||
			organization.Slug == "" {
			return ErrEvidenceMissing
		}
	}
	return nil
}

type SelfUpdateCommand struct {
	OperationID                  uuid.UUID
	RequestSHA256                [32]byte
	ExpectedSourceSchemaID       uuid.UUID
	ExpectedRevision             int64
	Target                       Document
	ActorID                      uuid.UUID
	ExpectedTokenVersion         int64
	ExpectedAuthorizationVersion int64
}

func (command SelfUpdateCommand) Validate() error {
	document := command.Target.CanonicalJSON()
	if !stableID(command.OperationID) || command.RequestSHA256 == [32]byte{} ||
		!stableID(command.ExpectedSourceSchemaID) || command.ExpectedRevision < 1 ||
		command.Target.Resource() != ResourceUser || !stableID(command.Target.SchemaID()) ||
		len(document) < 2 || len(document) > MaximumDocumentBytes ||
		command.Target.SHA256() == [32]byte{} ||
		(SelfEvidence{
			ActorID:                      command.ActorID,
			ExpectedTokenVersion:         command.ExpectedTokenVersion,
			ExpectedAuthorizationVersion: command.ExpectedAuthorizationVersion,
		}).Validate() != nil {
		return ErrInvalidCommand
	}
	return nil
}

type SelfPort interface {
	GetSelf(context.Context, transaction.TenantScope, SelfEvidence) (SelfRecord, error)
	UpdateSelf(
		context.Context, transaction.TenantScope, SelfUpdateCommand,
	) (MutationResult, error)
}
