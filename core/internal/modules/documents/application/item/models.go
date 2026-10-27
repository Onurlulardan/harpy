package item

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type Version struct {
	ID                           domain.ID
	Number                       uint64
	OriginalName, MIME, FileKind string
	PlaintextBytes               int64
	PlaintextSHA256              [32]byte
	CreatedAt                    time.Time
}

type Snapshot struct {
	ID, TenantID, FolderID, OwnerID, CurrentVersionID domain.ID
	State                                             domain.ItemState
	RowVersion                                        uint64
	Description                                       string
	Tags                                              []string
	Current                                           Version
	TrashedAt, RetentionEndsAt                        time.Time
	PurgeRequestedAt, PurgeStartsAt                   time.Time
	CreatedAt, UpdatedAt                              time.Time
}

type Cursor struct {
	UpdatedAt time.Time
	ID        domain.ID
}

type Page struct {
	Items []Snapshot
	Next  Cursor
}

type VersionPage struct {
	Items           []Version
	NextAfterNumber uint64
}

type GetRequest struct{ ItemID domain.ID }
type ListRequest struct {
	FolderID domain.ID
	After    Cursor
	Limit    uint16
}

type ListVersionsRequest struct {
	ItemID             domain.ID
	AfterVersionNumber uint64
	Limit              uint16
}

type MetadataRequest struct {
	ItemID         domain.ID
	RowVersion     uint64
	Description    string
	Tags           []string
	IdempotencyKey string
}

type MoveRequest struct {
	ItemID, FolderID domain.ID
	RowVersion       uint64
	IdempotencyKey   string
}

type MutationRequest struct {
	ItemID         domain.ID
	RowVersion     uint64
	IdempotencyKey string
}

type PurgeRequest struct {
	MutationRequest
	Reason string
}

type Mutation struct {
	ItemID, ActorID, TenantID, FolderID domain.ID
	OperationID, AuditID                domain.ID
	ExpectedRowVersion                  uint64
	Description                         string
	TagsJSON                            []byte
	Reason, IdempotencyKey              string
	RequestSHA256                       [32]byte
	OccurredAt, StepUpAt                time.Time
}
