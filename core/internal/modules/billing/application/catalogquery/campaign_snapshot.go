package catalogquery

import (
	"time"

	"github.com/google/uuid"
)

type CampaignSnapshot struct {
	id          uuid.UUID
	legacyID    uuid.UUID
	createdBy   uuid.UUID
	key         string
	revision    int64
	versionID   uuid.UUID
	version     int64
	state       string
	name        string
	description string
	bannerText  string
	autoApply   bool
	available   bool
	priority    int32
	startsAt    time.Time
	endsAt      *time.Time
	createdAt   time.Time
	updatedAt   time.Time
	discounts   []DiscountSnapshot
}

func (snapshot CampaignSnapshot) ID() uuid.UUID        { return snapshot.id }
func (snapshot CampaignSnapshot) LegacyID() uuid.UUID  { return snapshot.legacyID }
func (snapshot CampaignSnapshot) CreatedBy() uuid.UUID { return snapshot.createdBy }
func (snapshot CampaignSnapshot) Key() string          { return snapshot.key }
func (snapshot CampaignSnapshot) Revision() int64      { return snapshot.revision }
func (snapshot CampaignSnapshot) VersionID() uuid.UUID { return snapshot.versionID }
func (snapshot CampaignSnapshot) Version() int64       { return snapshot.version }
func (snapshot CampaignSnapshot) State() string        { return snapshot.state }
func (snapshot CampaignSnapshot) Name() string         { return snapshot.name }
func (snapshot CampaignSnapshot) Description() string  { return snapshot.description }
func (snapshot CampaignSnapshot) BannerText() string   { return snapshot.bannerText }
func (snapshot CampaignSnapshot) AutoApply() bool      { return snapshot.autoApply }
func (snapshot CampaignSnapshot) Available() bool      { return snapshot.available }
func (snapshot CampaignSnapshot) Priority() int32      { return snapshot.priority }
func (snapshot CampaignSnapshot) StartsAt() time.Time  { return snapshot.startsAt }
func (snapshot CampaignSnapshot) EndsAt() *time.Time   { return copyTime(snapshot.endsAt) }
func (snapshot CampaignSnapshot) CreatedAt() time.Time { return snapshot.createdAt }
func (snapshot CampaignSnapshot) UpdatedAt() time.Time { return snapshot.updatedAt }

func (snapshot CampaignSnapshot) Discounts() []DiscountSnapshot {
	result := make([]DiscountSnapshot, len(snapshot.discounts))
	for index := range snapshot.discounts {
		result[index] = snapshot.discounts[index].clone()
	}
	return result
}

type DiscountSnapshot struct {
	id                      uuid.UUID
	legacyID                uuid.UUID
	createdBy               uuid.UUID
	code                    string
	revision                int64
	versionID               uuid.UUID
	version                 int64
	state                   string
	name                    string
	description             string
	available               bool
	kind                    string
	basisPoints             *int64
	fixedMinorUnits         *int64
	currency                *string
	currencyRegistryVersion *int64
	duration                string
	durationCycles          *int64
	priority                int32
	allOfferings            bool
	offeringVersionIDs      []uuid.UUID
	stableOfferingIDs       []uuid.UUID
	legacyOfferingIDs       []uuid.UUID
	usedCount               int64
	campaignIDs             []uuid.UUID
	legacyCampaignIDs       []uuid.UUID
	validStart              time.Time
	validEnd                *time.Time
	globalUseLimit          *int64
	perTenantUseLimit       *int64
	createdAt               time.Time
	updatedAt               time.Time
}

func (snapshot DiscountSnapshot) ID() uuid.UUID           { return snapshot.id }
func (snapshot DiscountSnapshot) LegacyID() uuid.UUID     { return snapshot.legacyID }
func (snapshot DiscountSnapshot) CreatedBy() uuid.UUID    { return snapshot.createdBy }
func (snapshot DiscountSnapshot) Code() string            { return snapshot.code }
func (snapshot DiscountSnapshot) Revision() int64         { return snapshot.revision }
func (snapshot DiscountSnapshot) VersionID() uuid.UUID    { return snapshot.versionID }
func (snapshot DiscountSnapshot) Version() int64          { return snapshot.version }
func (snapshot DiscountSnapshot) State() string           { return snapshot.state }
func (snapshot DiscountSnapshot) Name() string            { return snapshot.name }
func (snapshot DiscountSnapshot) Description() string     { return snapshot.description }
func (snapshot DiscountSnapshot) Available() bool         { return snapshot.available }
func (snapshot DiscountSnapshot) Kind() string            { return snapshot.kind }
func (snapshot DiscountSnapshot) BasisPoints() *int64     { return copyInt64(snapshot.basisPoints) }
func (snapshot DiscountSnapshot) FixedMinorUnits() *int64 { return copyInt64(snapshot.fixedMinorUnits) }
func (snapshot DiscountSnapshot) Currency() *string       { return copyString(snapshot.currency) }
func (snapshot DiscountSnapshot) CurrencyRegistryVersion() *int64 {
	return copyInt64(snapshot.currencyRegistryVersion)
}
func (snapshot DiscountSnapshot) Duration() string       { return snapshot.duration }
func (snapshot DiscountSnapshot) DurationCycles() *int64 { return copyInt64(snapshot.durationCycles) }
func (snapshot DiscountSnapshot) Priority() int32        { return snapshot.priority }
func (snapshot DiscountSnapshot) AllOfferings() bool     { return snapshot.allOfferings }
func (snapshot DiscountSnapshot) OfferingVersionIDs() []uuid.UUID {
	return append([]uuid.UUID(nil), snapshot.offeringVersionIDs...)
}
func (snapshot DiscountSnapshot) StableOfferingIDs() []uuid.UUID {
	return append([]uuid.UUID(nil), snapshot.stableOfferingIDs...)
}
func (snapshot DiscountSnapshot) LegacyOfferingIDs() []uuid.UUID {
	return append([]uuid.UUID(nil), snapshot.legacyOfferingIDs...)
}
func (snapshot DiscountSnapshot) UsedCount() int64 { return snapshot.usedCount }
func (snapshot DiscountSnapshot) CampaignIDs() []uuid.UUID {
	return append([]uuid.UUID(nil), snapshot.campaignIDs...)
}
func (snapshot DiscountSnapshot) LegacyCampaignIDs() []uuid.UUID {
	return append([]uuid.UUID(nil), snapshot.legacyCampaignIDs...)
}
func (snapshot DiscountSnapshot) ValidStart() time.Time  { return snapshot.validStart }
func (snapshot DiscountSnapshot) ValidEnd() *time.Time   { return copyTime(snapshot.validEnd) }
func (snapshot DiscountSnapshot) GlobalUseLimit() *int64 { return copyInt64(snapshot.globalUseLimit) }
func (snapshot DiscountSnapshot) PerTenantUseLimit() *int64 {
	return copyInt64(snapshot.perTenantUseLimit)
}
func (snapshot DiscountSnapshot) CreatedAt() time.Time { return snapshot.createdAt }
func (snapshot DiscountSnapshot) UpdatedAt() time.Time { return snapshot.updatedAt }

func (snapshot DiscountSnapshot) clone() DiscountSnapshot {
	snapshot.basisPoints = copyInt64(snapshot.basisPoints)
	snapshot.fixedMinorUnits = copyInt64(snapshot.fixedMinorUnits)
	snapshot.currency = copyString(snapshot.currency)
	snapshot.currencyRegistryVersion = copyInt64(snapshot.currencyRegistryVersion)
	snapshot.durationCycles = copyInt64(snapshot.durationCycles)
	snapshot.offeringVersionIDs = append([]uuid.UUID(nil), snapshot.offeringVersionIDs...)
	snapshot.stableOfferingIDs = append([]uuid.UUID(nil), snapshot.stableOfferingIDs...)
	snapshot.legacyOfferingIDs = append([]uuid.UUID(nil), snapshot.legacyOfferingIDs...)
	snapshot.campaignIDs = append([]uuid.UUID(nil), snapshot.campaignIDs...)
	snapshot.legacyCampaignIDs = append([]uuid.UUID(nil), snapshot.legacyCampaignIDs...)
	snapshot.validEnd = copyTime(snapshot.validEnd)
	snapshot.globalUseLimit = copyInt64(snapshot.globalUseLimit)
	snapshot.perTenantUseLimit = copyInt64(snapshot.perTenantUseLimit)
	return snapshot
}
