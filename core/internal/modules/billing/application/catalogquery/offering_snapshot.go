package catalogquery

import (
	"time"

	"github.com/google/uuid"
)

type OfferingSnapshot struct {
	id              uuid.UUID
	key             string
	revision        int64
	versionID       uuid.UUID
	version         int64
	state           string
	public          bool
	available       bool
	name            string
	description     string
	sortOrder       int32
	highlightLabel  string
	color           string
	billingInterval string
	base            bool
	trialDays       int32
	graceDays       int32
	effectiveStart  time.Time
	effectiveEnd    *time.Time
	createdAt       time.Time
	updatedAt       time.Time
	prices          []PriceSnapshot
	entitlements    []EntitlementSnapshot
}

func (snapshot OfferingSnapshot) ID() uuid.UUID             { return snapshot.id }
func (snapshot OfferingSnapshot) Key() string               { return snapshot.key }
func (snapshot OfferingSnapshot) Revision() int64           { return snapshot.revision }
func (snapshot OfferingSnapshot) VersionID() uuid.UUID      { return snapshot.versionID }
func (snapshot OfferingSnapshot) Version() int64            { return snapshot.version }
func (snapshot OfferingSnapshot) Name() string              { return snapshot.name }
func (snapshot OfferingSnapshot) Description() string       { return snapshot.description }
func (snapshot OfferingSnapshot) SortOrder() int32          { return snapshot.sortOrder }
func (snapshot OfferingSnapshot) HighlightLabel() string    { return snapshot.highlightLabel }
func (snapshot OfferingSnapshot) Color() string             { return snapshot.color }
func (snapshot OfferingSnapshot) BillingInterval() string   { return snapshot.billingInterval }
func (snapshot OfferingSnapshot) State() string             { return snapshot.state }
func (snapshot OfferingSnapshot) Public() bool              { return snapshot.public }
func (snapshot OfferingSnapshot) Available() bool           { return snapshot.available }
func (snapshot OfferingSnapshot) Base() bool                { return snapshot.base }
func (snapshot OfferingSnapshot) TrialDays() int32          { return snapshot.trialDays }
func (snapshot OfferingSnapshot) GraceDays() int32          { return snapshot.graceDays }
func (snapshot OfferingSnapshot) EffectiveStart() time.Time { return snapshot.effectiveStart }
func (snapshot OfferingSnapshot) CreatedAt() time.Time      { return snapshot.createdAt }
func (snapshot OfferingSnapshot) UpdatedAt() time.Time      { return snapshot.updatedAt }

func (snapshot OfferingSnapshot) EffectiveEnd() *time.Time {
	return copyTime(snapshot.effectiveEnd)
}

func (snapshot OfferingSnapshot) Prices() []PriceSnapshot {
	return append([]PriceSnapshot(nil), snapshot.prices...)
}

func (snapshot OfferingSnapshot) Entitlements() []EntitlementSnapshot {
	return append([]EntitlementSnapshot(nil), snapshot.entitlements...)
}

type PriceSnapshot struct {
	id                      uuid.UUID
	cadence                 string
	minorUnits              int64
	currency                string
	currencyRegistryVersion int64
}

func (snapshot PriceSnapshot) ID() uuid.UUID     { return snapshot.id }
func (snapshot PriceSnapshot) Cadence() string   { return snapshot.cadence }
func (snapshot PriceSnapshot) MinorUnits() int64 { return snapshot.minorUnits }
func (snapshot PriceSnapshot) Currency() string  { return snapshot.currency }
func (snapshot PriceSnapshot) CurrencyRegistryVersion() int64 {
	return snapshot.currencyRegistryVersion
}

type EntitlementSnapshot struct {
	key       string
	kind      string
	boolean   *bool
	integer   *int64
	unlimited bool
}

func (snapshot EntitlementSnapshot) Key() string     { return snapshot.key }
func (snapshot EntitlementSnapshot) Kind() string    { return snapshot.kind }
func (snapshot EntitlementSnapshot) Unlimited() bool { return snapshot.unlimited }
func (snapshot EntitlementSnapshot) Boolean() *bool  { return copyBool(snapshot.boolean) }
func (snapshot EntitlementSnapshot) Integer() *int64 { return copyInt64(snapshot.integer) }

type OfferingPage struct {
	items      []OfferingSnapshot
	nextCursor string
}

func (page OfferingPage) Items() []OfferingSnapshot {
	return append([]OfferingSnapshot(nil), page.items...)
}

func (page OfferingPage) NextCursor() string { return page.nextCursor }
