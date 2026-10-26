package ports

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

type Clock interface{ Now() time.Time }
type IDGenerator interface {
	New(time.Time) (domain.ID, error)
}

type Authorizer interface {
	Authorize(context.Context, Scope, string, domain.ID) error
}

type PrincipalSnapshot struct {
	Principals         []domain.Principal
	PermissionRevision int64
}

type PrincipalResolver interface {
	Resolve(context.Context, Scope) (PrincipalSnapshot, error)
	Exists(context.Context, Scope, domain.Principal) (bool, int64, error)
}

type ResourceAuthorityRequest struct {
	Owner         domain.SourceOwner
	Kind          string
	ID, VersionID domain.ID
	Epoch         int64
}

type ResourceAuthority interface {
	AuthorizeRead(context.Context, Scope, ResourceAuthorityRequest) (bool, int64, error)
}

type ResourceReferenceCommand struct {
	Link        domain.ResourceLink
	OperationID domain.ID
	OccurredAt  time.Time
}

type ResourceReferenceReceipt struct {
	ReferenceID, OperationID  domain.ID
	TargetID, TargetVersionID domain.ID
	ReferenceEpoch            int64
}

func (receipt ResourceReferenceReceipt) Matches(
	command ResourceReferenceCommand,
) bool {
	return receipt.ReferenceID == command.Link.ID &&
		receipt.OperationID == command.OperationID &&
		receipt.TargetID == command.Link.TargetID &&
		receipt.TargetVersionID == command.Link.TargetVersionID &&
		receipt.ReferenceEpoch > 0
}

type ResourceReferenceRegistry interface {
	RegisterRead(context.Context, Scope, ResourceReferenceCommand) (
		ResourceReferenceReceipt, error)
	ReleaseRead(context.Context, Scope, ResourceReferenceCommand) (
		ResourceReferenceReceipt, error)
}

type TZDBRegistry interface {
	Load(context.Context, string, string) (*time.Location, error)
}

type NotificationIntent struct {
	DedupeKey, Kind, SubjectID, RecipientID string
	NotBefore                               time.Time
}

type NotificationPort interface {
	RecordIntent(context.Context, Scope, NotificationIntent) error
}

type EventIntent struct {
	EventID, AggregateID domain.ID
	Type, AggregateKind  string
	AggregateVersion     int64
	EventSequence        int64
	OccurredAt           time.Time
	IdempotencyKey       string
	Payload              []byte
}

type EventPublisher interface {
	Record(context.Context, Scope, EventIntent) error
}
