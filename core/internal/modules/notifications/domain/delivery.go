package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrDeliveryInvalid  = errors.New("notification delivery is invalid")
	ErrDeliveryConflict = errors.New("notification delivery transition conflicts")
)

type DeliveryState string

const (
	DeliveryPending    DeliveryState = "pending"
	DeliverySuppressed DeliveryState = "suppressed"
	DeliverySending    DeliveryState = "sending"
	DeliveryAccepted   DeliveryState = "accepted"
	DeliveryDelivered  DeliveryState = "delivered"
	DeliveryBounced    DeliveryState = "bounced"
	DeliveryFailed     DeliveryState = "failed"
	DeliveryCancelled  DeliveryState = "cancelled"
)

type Delivery struct {
	ID, MessageID, ProviderConfigID ID
	Channel                         Channel
	State                           DeliveryState
	Attempt, MaximumAttempts        int
	LeaseOwner                      string
	LeaseGeneration                 int64
	LeaseUntil, NotBefore           time.Time
	OutcomeCode                     string
	OutcomeUnknown                  bool
	Mandatory                       bool
	Revision                        int64
}

func NewDelivery(value Delivery) (Delivery, error) {
	if !value.ID.IsV7() || !value.MessageID.Valid() || !value.Channel.Valid() ||
		value.Channel == ChannelInApp || value.State != DeliveryPending ||
		value.Attempt != 0 || value.MaximumAttempts != 8 || value.NotBefore.IsZero() ||
		value.Revision != 1 {
		return Delivery{}, ErrDeliveryInvalid
	}
	value.NotBefore = value.NotBefore.UTC()
	return value, nil
}

func NewSuppressedDelivery(value Delivery, code string) (Delivery, error) {
	value.State = DeliveryPending
	pending, err := NewDelivery(value)
	if err != nil {
		return Delivery{}, err
	}
	return pending.Suppress(code)
}

func NewFailedDelivery(value Delivery, code string) (Delivery, error) {
	value.State = DeliveryPending
	pending, err := NewDelivery(value)
	if err != nil || !validOutcomeCode(code) {
		return Delivery{}, ErrDeliveryInvalid
	}
	pending.State, pending.OutcomeCode = DeliveryFailed, code
	pending.Revision++
	return pending, nil
}

func (value Delivery) Claim(owner string, now time.Time) (Delivery, error) {
	owner = strings.TrimSpace(owner)
	if value.State != DeliveryPending || value.OutcomeUnknown || owner == "" ||
		len(owner) > 128 || now.IsZero() || now.Before(value.NotBefore) ||
		value.Attempt >= value.MaximumAttempts {
		return Delivery{}, ErrDeliveryConflict
	}
	value.State, value.LeaseOwner = DeliverySending, owner
	value.Attempt++
	value.LeaseGeneration++
	value.LeaseUntil = now.UTC().Add(time.Minute)
	value.Revision++
	return value, nil
}

func (value Delivery) Accept(owner string, generation int64, code string) (Delivery, error) {
	if !value.matchesLease(owner, generation) || !validOutcomeCode(code) {
		return Delivery{}, ErrDeliveryConflict
	}
	value.State, value.OutcomeCode = DeliveryAccepted, code
	value.clearLease()
	value.Revision++
	return value, nil
}

func (value Delivery) Fail(owner string, generation int64, code string,
	outcomeUnknown bool,
) (Delivery, error) {
	if !value.matchesLease(owner, generation) || !validOutcomeCode(code) {
		return Delivery{}, ErrDeliveryConflict
	}
	value.State, value.OutcomeCode = DeliveryFailed, code
	value.OutcomeUnknown = outcomeUnknown
	value.clearLease()
	value.Revision++
	return value, nil
}

func (value Delivery) RecordProviderReceipt(delivered bool, code string) (Delivery, error) {
	if value.State != DeliveryAccepted || !validOutcomeCode(code) {
		return Delivery{}, ErrDeliveryConflict
	}
	if delivered {
		value.State = DeliveryDelivered
	} else {
		value.State = DeliveryBounced
	}
	value.OutcomeCode = code
	value.Revision++
	return value, nil
}

func (value Delivery) Retry(now, next time.Time) (Delivery, error) {
	if value.State != DeliveryFailed || value.OutcomeUnknown || now.IsZero() ||
		next.Before(now) || value.Attempt >= value.MaximumAttempts {
		return Delivery{}, ErrDeliveryConflict
	}
	value.State, value.NotBefore, value.OutcomeCode = DeliveryPending, next.UTC(), ""
	value.Revision++
	return value, nil
}

func (value Delivery) Suppress(code string) (Delivery, error) {
	if value.State != DeliveryPending || !validOutcomeCode(code) {
		return Delivery{}, ErrDeliveryConflict
	}
	value.State, value.OutcomeCode = DeliverySuppressed, code
	value.Revision++
	return value, nil
}

func (value Delivery) matchesLease(owner string, generation int64) bool {
	return value.State == DeliverySending && owner != "" && owner == value.LeaseOwner &&
		generation > 0 && generation == value.LeaseGeneration
}

func (value *Delivery) clearLease() {
	value.LeaseOwner, value.LeaseUntil = "", time.Time{}
}

func validOutcomeCode(value string) bool {
	if len(value) < 2 || len(value) > 64 || value[0] < 'A' || value[0] > 'Z' {
		return false
	}
	for _, character := range value[1:] {
		if character != '_' && (character < 'A' || character > 'Z') &&
			(character < '0' || character > '9') {
			return false
		}
	}
	return true
}
