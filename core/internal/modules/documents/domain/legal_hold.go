package domain

import (
	"errors"
	"strings"
	"time"
)

var ErrLegalHoldInvalid = errors.New("documents legal hold input is invalid")

type LegalHoldState string

const (
	LegalHoldActive   LegalHoldState = "active"
	LegalHoldReleased LegalHoldState = "released"
)

type LegalHold struct {
	ID, TenantID, ItemID      ID
	SourceOwner, SourceCaseID string
	EvidenceSHA256            [32]byte
	State                     LegalHoldState
	PlacedAt, ReleasedAt      time.Time
}

func NewLegalHold(id, tenantID, itemID ID, owner, caseID string,
	evidence [32]byte, now time.Time,
) (LegalHold, error) {
	caseID = strings.TrimSpace(caseID)
	if !id.UUIDv7() || !tenantID.Valid() || !itemID.UUIDv7() ||
		!validLegalHoldOwner(owner) || len(caseID) < 1 || len([]byte(caseID)) > 256 ||
		evidence == [32]byte{} || !validInstant(now) {
		return LegalHold{}, ErrLegalHoldInvalid
	}
	return LegalHold{ID: id, TenantID: tenantID, ItemID: itemID,
		SourceOwner: owner, SourceCaseID: caseID, EvidenceSHA256: evidence,
		State: LegalHoldActive, PlacedAt: now.UTC()}, nil
}

func (hold LegalHold) Release(now time.Time) (LegalHold, error) {
	if hold.State != LegalHoldActive || !validInstant(now) {
		return LegalHold{}, ErrLegalHoldInvalid
	}
	hold.State, hold.ReleasedAt = LegalHoldReleased, now.UTC()
	return hold, nil
}

func validLegalHoldOwner(value string) bool {
	switch value {
	case "BILLING", "FORMS", "CATALOG", "IAM":
		return true
	default:
		return false
	}
}
