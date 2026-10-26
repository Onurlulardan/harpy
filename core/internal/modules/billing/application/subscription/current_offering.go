package subscriptionapp

import (
	"time"

	"github.com/google/uuid"
)

type CurrentOfferingInput struct {
	ID, VersionID                          uuid.UUID
	Key, State, Name, Description          string
	HighlightLabel, Color, BillingInterval string
	Revision, Version                      int64
	SortOrder                              int32
	Public, Base                           bool
	TrialDays, GraceDays                   int32
	ExpectedPriceCount, ExpectedGrantCount int
	EffectiveStart, CreatedAt, UpdatedAt   time.Time
	EffectiveEnd                           *time.Time
	Prices                                 []CurrentOfferingPrice
	Entitlements                           []CurrentOfferingEntitlement
}

type CurrentOfferingPrice struct {
	ID                      uuid.UUID
	Cadence                 string
	MinorUnits              int64
	Currency                string
	CurrencyRegistryVersion int64
}

type CurrentOfferingEntitlement struct {
	Key       string
	Kind      string
	Boolean   *bool
	Integer   *int64
	Unlimited bool
}

type CurrentOffering struct {
	id, versionID                          uuid.UUID
	key, state, name, description          string
	highlightLabel, color, billingInterval string
	revision, version                      int64
	sortOrder                              int32
	public, base                           bool
	trialDays, graceDays                   int32
	effectiveStart, createdAt, updatedAt   time.Time
	effectiveEnd                           *time.Time
	prices                                 []CurrentOfferingPrice
	entitlements                           []CurrentOfferingEntitlement
}

func NewCurrentOffering(input CurrentOfferingInput) (CurrentOffering, error) {
	value := CurrentOffering{
		id: input.ID, versionID: input.VersionID, key: input.Key,
		state: input.State, name: input.Name, description: input.Description,
		highlightLabel: input.HighlightLabel, color: input.Color,
		billingInterval: input.BillingInterval, revision: input.Revision,
		version: input.Version, sortOrder: input.SortOrder, public: input.Public,
		base: input.Base, trialDays: input.TrialDays, graceDays: input.GraceDays,
		effectiveStart: input.EffectiveStart, effectiveEnd: copiedInstant(input.EffectiveEnd),
		createdAt: input.CreatedAt, updatedAt: input.UpdatedAt,
		prices:       copyOfferingPrices(input.Prices),
		entitlements: copyOfferingEntitlements(input.Entitlements),
	}
	if !value.valid(input.ExpectedPriceCount, input.ExpectedGrantCount) {
		return CurrentOffering{}, ErrCurrentSubscriptionInvariant
	}
	return value, nil
}

func (value CurrentOffering) ID() uuid.UUID             { return value.id }
func (value CurrentOffering) VersionID() uuid.UUID      { return value.versionID }
func (value CurrentOffering) Key() string               { return value.key }
func (value CurrentOffering) State() string             { return value.state }
func (value CurrentOffering) Name() string              { return value.name }
func (value CurrentOffering) Description() string       { return value.description }
func (value CurrentOffering) HighlightLabel() string    { return value.highlightLabel }
func (value CurrentOffering) Color() string             { return value.color }
func (value CurrentOffering) BillingInterval() string   { return value.billingInterval }
func (value CurrentOffering) Revision() int64           { return value.revision }
func (value CurrentOffering) Version() int64            { return value.version }
func (value CurrentOffering) SortOrder() int32          { return value.sortOrder }
func (value CurrentOffering) Public() bool              { return value.public }
func (value CurrentOffering) Base() bool                { return value.base }
func (value CurrentOffering) TrialDays() int32          { return value.trialDays }
func (value CurrentOffering) GraceDays() int32          { return value.graceDays }
func (value CurrentOffering) EffectiveStart() time.Time { return value.effectiveStart }
func (value CurrentOffering) EffectiveEnd() *time.Time  { return copiedInstant(value.effectiveEnd) }
func (value CurrentOffering) CreatedAt() time.Time      { return value.createdAt }
func (value CurrentOffering) UpdatedAt() time.Time      { return value.updatedAt }
func (value CurrentOffering) Prices() []CurrentOfferingPrice {
	return copyOfferingPrices(value.prices)
}
func (value CurrentOffering) Entitlements() []CurrentOfferingEntitlement {
	return copyOfferingEntitlements(value.entitlements)
}
