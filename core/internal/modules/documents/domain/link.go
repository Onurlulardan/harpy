package domain

import (
	"errors"
	"time"
)

var ErrLinkInvalid = errors.New("documents link input is invalid")

type LinkState string

const (
	LinkActive   LinkState = "active"
	LinkReleased LinkState = "released"
)

type Link struct {
	ID, TenantID, ItemID, SourceID, SourceVersionID ID
	SourceOwner, SourceKind, Usage                  string
	ReferenceEpoch                                  uint64
	State                                           LinkState
	CreatedAt, ReleasedAt                           time.Time
}

func NewLink(id, tenantID, itemID, sourceID, sourceVersionID ID,
	owner, kind, usage string, epoch uint64, now time.Time,
) (Link, error) {
	if !id.UUIDv7() || !tenantID.Valid() || !itemID.UUIDv7() || !sourceID.Valid() ||
		!sourceVersionID.Valid() || !validLinkOwner(owner) || kind == "" || len(kind) > 128 ||
		usage != "READ" && usage != "LIFECYCLE" || epoch == 0 || !validInstant(now) {
		return Link{}, ErrLinkInvalid
	}
	return Link{ID: id, TenantID: tenantID, ItemID: itemID, SourceID: sourceID,
		SourceVersionID: sourceVersionID, SourceOwner: owner, SourceKind: kind,
		Usage: usage, ReferenceEpoch: epoch, State: LinkActive, CreatedAt: now.UTC()}, nil
}

func (link Link) Release(now time.Time) (Link, error) {
	if link.State != LinkActive || !validInstant(now) {
		return Link{}, ErrLinkInvalid
	}
	link.State, link.ReleasedAt = LinkReleased, now.UTC()
	return link, nil
}

func validLinkOwner(value string) bool {
	switch value {
	case "FORMS", "WORKFLOW", "WORK", "CATALOG", "IAM":
		return true
	default:
		return false
	}
}
