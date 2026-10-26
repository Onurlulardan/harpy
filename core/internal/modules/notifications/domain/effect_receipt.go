package domain

import (
	"crypto/sha256"
	"errors"
	"strings"
	"time"
)

var (
	ErrEffectInvalid = errors.New("notification effect identity is invalid")
	ErrEffectReused  = errors.New("notification effect key was reused with different content")
)

type EffectOutcome string

const (
	EffectPending        EffectOutcome = "pending"
	EffectCreated        EffectOutcome = "created"
	EffectSuppressed     EffectOutcome = "suppressed"
	EffectOutcomeUnknown EffectOutcome = "outcome_unknown"
)

type EffectCandidate struct {
	Producer, EffectKey string
	RecipientID         ID
	CanonicalDigest     [32]byte
}

func (value EffectCandidate) Valid() bool {
	value.Producer = strings.TrimSpace(value.Producer)
	value.EffectKey = strings.TrimSpace(value.EffectKey)
	return value.Producer != "" && len(value.Producer) <= 64 &&
		value.EffectKey != "" && len(value.EffectKey) <= 512 &&
		value.RecipientID.Valid() && value.CanonicalDigest != [32]byte{}
}

func (value EffectCandidate) KeyHash() [32]byte {
	return sha256.Sum256([]byte(value.EffectKey))
}

type EffectReceipt struct {
	Producer               string
	EffectKeyHash          [32]byte
	RecipientID            ID
	CanonicalDigest        [32]byte
	MessageID              ID
	LegacyReceiptReference string
	Outcome                EffectOutcome
	CreatedAt, RetainUntil time.Time
}

func NewEffectReceipt(candidate EffectCandidate, messageID ID,
	createdAt, retainUntil time.Time,
) (EffectReceipt, error) {
	if !candidate.Valid() || !messageID.Valid() || createdAt.IsZero() ||
		retainUntil.Before(createdAt.AddDate(1, 0, 0)) {
		return EffectReceipt{}, ErrEffectInvalid
	}
	return EffectReceipt{Producer: candidate.Producer,
		EffectKeyHash: candidate.KeyHash(), RecipientID: candidate.RecipientID,
		CanonicalDigest: candidate.CanonicalDigest, MessageID: messageID,
		Outcome: EffectCreated, CreatedAt: createdAt.UTC(), RetainUntil: retainUntil.UTC()}, nil
}

func (value EffectReceipt) Replay(candidate EffectCandidate) (ID, EffectOutcome, error) {
	if !candidate.Valid() || candidate.Producer != value.Producer ||
		candidate.RecipientID != value.RecipientID ||
		candidate.KeyHash() != value.EffectKeyHash {
		return ID{}, "", ErrEffectInvalid
	}
	if candidate.CanonicalDigest != value.CanonicalDigest {
		return ID{}, "", ErrEffectReused
	}
	return value.MessageID, value.Outcome, nil
}
