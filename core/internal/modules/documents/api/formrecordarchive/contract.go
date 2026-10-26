package formrecordarchive

import (
	"context"
	"time"
)

const (
	ContractVersion  = "documents-form-record-archive.v1"
	MaximumItemCount = 1000
)

type CreateRequest struct {
	DefinitionID, RecordID ID
	ItemIDs                []ID
	RetryOfArchiveID       ID
	IdempotencyKey         string
}

func (request CreateRequest) Valid() bool {
	return sourceItemsValid(request.DefinitionID, request.RecordID,
		request.ItemIDs) && (!request.RetryOfArchiveID.Valid() ||
		request.RetryOfArchiveID.UUIDv7()) && len(request.IdempotencyKey) >= 8 &&
		len(request.IdempotencyKey) <= 256
}

type GetRequest struct {
	ArchiveID, RetryOfArchiveID ID
	ItemIDs                     []ID
	RequestSHA256               [32]byte
}

func (request GetRequest) Valid() bool {
	return request.ArchiveID.UUIDv7() && validItems(request.ItemIDs) &&
		(!request.RetryOfArchiveID.Valid() || request.RetryOfArchiveID.UUIDv7()) &&
		request.RequestSHA256 != [32]byte{}
}

type State string

const (
	StatePending   State = "pending"
	StateCancelled State = "cancelled"
	StateReady     State = "ready"
	StateExpired   State = "expired"
	StatePurged    State = "purged"
)

type Snapshot struct {
	ArchiveID, JobID, RetryOfArchiveID ID
	State                              State
	RequestSHA256                      [32]byte
	EntryCount                         int
	PlaintextBytes, CiphertextBytes    int64
	ReadyAt, ExpiresAt                 time.Time
}

type Owner interface {
	Create(context.Context, Access, CreateRequest) (Snapshot, error)
	Get(context.Context, Access, GetRequest) (Snapshot, error)
}

func sourceItemsValid(definitionID, recordID ID, items []ID) bool {
	return definitionID.UUIDv7() && recordID.UUIDv7() && validItems(items)
}

func validItems(items []ID) bool {
	if len(items) == 0 || len(items) > MaximumItemCount {
		return false
	}
	seen := make(map[ID]struct{}, len(items))
	for _, itemID := range items {
		if !itemID.UUIDv7() {
			return false
		}
		if _, duplicate := seen[itemID]; duplicate {
			return false
		}
		seen[itemID] = struct{}{}
	}
	return true
}
