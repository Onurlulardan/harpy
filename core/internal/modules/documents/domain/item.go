package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrItemInvalid    = errors.New("documents item input is invalid")
	ErrItemTransition = errors.New("documents item transition is invalid")
	ErrLegalHold      = errors.New("documents item is under legal hold")
)

const (
	TrashRetention   = 30 * 24 * time.Hour
	PurgeCooling     = 24 * time.Hour
	TombstoneKeep    = 365 * 24 * time.Hour
	StepUpMaximumAge = 10 * time.Minute
)

type ItemState string

const (
	ItemActive       ItemState = "active"
	ItemTrashed      ItemState = "trashed"
	ItemPurgeCooling ItemState = "purge_cooling"
	ItemPurging      ItemState = "purging"
	ItemPurged       ItemState = "purged"
)

type Item struct {
	ID, TenantID, FolderID, CurrentVersionID ID
	State                                    ItemState
	ResumeState                              ItemState
	TrashedAt, RetentionEndsAt               time.Time
	PurgeRequestedAt, PurgeStartsAt          time.Time
	PurgedAt, TombstoneUntil                 time.Time
	PurgeReason                              string
	PurgeEvidenceSHA256                      [32]byte
}

func NewItem(id, tenantID, folderID, currentVersionID ID) (Item, error) {
	if !id.UUIDv7() || !tenantID.Valid() || !folderID.UUIDv7() ||
		!currentVersionID.UUIDv7() {
		return Item{}, ErrItemInvalid
	}
	return Item{ID: id, TenantID: tenantID, FolderID: folderID,
		CurrentVersionID: currentVersionID, State: ItemActive}, nil
}

func (item Item) Trash(now time.Time) (Item, error) {
	if item.State != ItemActive || !validInstant(now) {
		return Item{}, ErrItemTransition
	}
	item.State, item.TrashedAt = ItemTrashed, now.UTC()
	item.RetentionEndsAt = item.TrashedAt.Add(TrashRetention)
	return item, nil
}

func (item Item) Restore(now time.Time) (Item, error) {
	if item.State != ItemTrashed || !validInstant(now) || !now.Before(item.RetentionEndsAt) {
		return Item{}, ErrItemTransition
	}
	item.State = ItemActive
	item.TrashedAt, item.RetentionEndsAt = time.Time{}, time.Time{}
	return item, nil
}

func (item Item) RequestPurge(now, stepUpAt time.Time, reason string) (Item, error) {
	trimmed := strings.TrimSpace(reason)
	if item.State != ItemActive && item.State != ItemTrashed || !validInstant(now) ||
		!validInstant(stepUpAt) || stepUpAt.After(now) || now.Sub(stepUpAt) > StepUpMaximumAge ||
		len(trimmed) < 3 || len([]byte(trimmed)) > 2048 {
		return Item{}, ErrItemTransition
	}
	item.ResumeState, item.State = item.State, ItemPurgeCooling
	item.PurgeRequestedAt, item.PurgeStartsAt = now.UTC(), now.UTC().Add(PurgeCooling)
	item.PurgeReason = trimmed
	return item, nil
}

func (item Item) CancelPurge(now time.Time) (Item, error) {
	if item.State != ItemPurgeCooling || !validInstant(now) || !now.Before(item.PurgeStartsAt) {
		return Item{}, ErrItemTransition
	}
	item.State, item.ResumeState = item.ResumeState, ""
	item.PurgeRequestedAt, item.PurgeStartsAt = time.Time{}, time.Time{}
	item.PurgeReason = ""
	return item, nil
}

func (item Item) BeginManualPurge(now time.Time, legalHold bool) (Item, error) {
	if legalHold {
		return Item{}, ErrLegalHold
	}
	if item.State != ItemPurgeCooling || !validInstant(now) || now.Before(item.PurgeStartsAt) {
		return Item{}, ErrItemTransition
	}
	item.State = ItemPurging
	return item, nil
}

func (item Item) BeginRetentionPurge(now time.Time, legalHold bool) (Item, error) {
	if legalHold {
		return Item{}, ErrLegalHold
	}
	if item.State != ItemTrashed || !validInstant(now) || now.Before(item.RetentionEndsAt) {
		return Item{}, ErrItemTransition
	}
	item.State = ItemPurging
	return item, nil
}

func (item Item) FinishPurge(now time.Time, pendingObjects int, evidence [32]byte) (Item, error) {
	if item.State != ItemPurging || !validInstant(now) || pendingObjects != 0 ||
		evidence == [32]byte{} {
		return Item{}, ErrItemTransition
	}
	item.State, item.PurgedAt = ItemPurged, now.UTC()
	item.TombstoneUntil = item.PurgedAt.Add(TombstoneKeep)
	item.PurgeEvidenceSHA256 = evidence
	return item, nil
}

func validInstant(value time.Time) bool { return !value.IsZero() }
