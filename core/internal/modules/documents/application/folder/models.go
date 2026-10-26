package folder

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type Page struct {
	Folders []domain.Folder
	NextID  domain.ID
}

type Folder = domain.Folder

type Mutation struct {
	OperationID, AuditID, ActorID, FolderID domain.ID
	ParentID                                domain.ID
	ExpectedVersion                         uint64
	Name, IdempotencyKey, Action            string
	RequestSHA256                           [32]byte
	OccurredAt                              time.Time
}

type CreateRequest struct {
	ParentID                    domain.ID
	DisplayName, IdempotencyKey string
}

type GetRequest struct{ FolderID domain.ID }

type ListRequest struct {
	ParentID domain.ID
	AfterID  domain.ID
	Limit    uint16
}

type UpdateRequest struct {
	FolderID                    domain.ID
	ExpectedVersion             uint64
	DisplayName, IdempotencyKey string
}

type MoveRequest struct {
	FolderID, ParentID domain.ID
	ExpectedVersion    uint64
	IdempotencyKey     string
}

type DeleteRequest struct {
	FolderID        domain.ID
	ExpectedVersion uint64
	IdempotencyKey  string
}

type RestoreRequest = DeleteRequest
