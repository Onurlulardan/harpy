package knowledgefile

import (
	"time"

	knowledgeapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/knowledgefile"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

const (
	OperationKnowledgeCreate = "ai.knowledge.create"
	OperationKnowledgeDelete = "ai.knowledge.delete"
	knowledgeFolderName      = "AI Knowledge Base"
	knowledgeFolderKey       = "ai-knowledge-folder-v1"
)

type Access struct {
	TenantID, ActorID    domain.ID
	AuthorizationVersion int64
}

type Lease struct {
	LeaseID, TenantID, ActorID  domain.ID
	UploadID, ItemID, VersionID domain.ID
	ReferenceID, SourceID       domain.ID
	ExpectedBytes               int64
	ExpectedSHA256              [32]byte
	Filename, MIME              string
	DetectedMIME, ScanRevision  string
	State                       string
	ItemRowVersion              uint64
	ScannedAt, ExpiresAt        time.Time
}

type StageSeed struct {
	Lease
	CreatedAt time.Time
}

type AdoptMutation struct {
	LeaseID, ItemID, VersionID, ReferenceID domain.ID
	SourceID                                domain.ID
	IdempotencyKey                          string
}

type ReleaseMutation struct {
	ItemID, ReferenceID, SourceID domain.ID
	IdempotencyKey                string
}

func stagedWire(value Lease) knowledgeapi.Staged {
	return knowledgeapi.Staged{LeaseID: knowledgeapi.ID(value.LeaseID),
		ItemID: knowledgeapi.ID(value.ItemID), VersionID: knowledgeapi.ID(value.VersionID),
		Filename: value.Filename, MIME: value.MIME, DetectedMIME: value.DetectedMIME,
		ScanRevision: value.ScanRevision, Bytes: value.ExpectedBytes,
		SHA256: value.ExpectedSHA256, Clean: value.State == "STAGED",
		ScannedAt: value.ScannedAt, ExpiresAt: value.ExpiresAt}
}
