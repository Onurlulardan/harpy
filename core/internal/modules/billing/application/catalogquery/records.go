package catalogquery

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Store interface {
	ListOfferings(context.Context, ListStoreQuery) ([]OfferingRecord, error)
	GetOffering(context.Context, GetStoreQuery) (OfferingRecord, bool, error)
	ReadActiveCampaign(context.Context, CampaignStoreQuery) (CampaignRecord, bool, error)
	ResolvePlatformCatalog(
		context.Context, PlatformResolveStoreQuery,
	) (PlatformCatalogRecord, bool, error)
	MaterializeLegacyIdentity(
		context.Context, LegacyIdentityStoreQuery,
	) (LegacyIdentityRecord, bool, error)
}

type ListStoreQuery struct {
	AfterKey string
	Limit    int
	Instant  time.Time
}

type GetStoreQuery struct {
	Key     string
	Instant time.Time
}

type CampaignStoreQuery struct{ Instant time.Time }

type PlatformCatalogKind string

const (
	PlatformOffering PlatformCatalogKind = "offering"
	PlatformDiscount PlatformCatalogKind = "discount"
	PlatformCampaign PlatformCatalogKind = "campaign"
)

type PlatformResolveStoreQuery struct {
	Kind PlatformCatalogKind
	ID   uuid.UUID
}

type LegacyIdentityStoreQuery struct {
	Kind     PlatformCatalogKind
	LegacyID uuid.UUID
}

type LegacyIdentityRecord struct {
	LegacyID    uuid.UUID
	CanonicalID uuid.UUID
	Kind        PlatformCatalogKind
}

type PlatformCatalogRecord struct {
	Offering *OfferingRecord
	Discount *DiscountRecord
	Campaign *CampaignRecord
}

type OfferingRecord struct {
	ID                 uuid.UUID
	Key                string
	Revision           int64
	VersionID          uuid.UUID
	Version            int64
	State              string
	Public             bool
	Available          bool
	Name               string
	Description        string
	SortOrder          int32
	HighlightLabel     string
	Color              string
	BillingInterval    string
	Base               bool
	TrialDays          int32
	GraceDays          int32
	ExpectedPriceCount int
	ExpectedGrantCount int
	EffectiveStart     time.Time
	EffectiveEnd       *time.Time
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Prices             []PriceRecord
	Entitlements       []EntitlementRecord
}

type PriceRecord struct {
	ID                      uuid.UUID
	Cadence                 string
	MinorUnits              int64
	Currency                string
	CurrencyRegistryVersion int64
}

type EntitlementRecord struct {
	Key       string
	Kind      string
	Boolean   *bool
	Integer   *int64
	Unlimited bool
}

type CampaignRecord struct {
	ID          uuid.UUID
	LegacyID    uuid.UUID
	CreatedBy   uuid.UUID
	Key         string
	Revision    int64
	VersionID   uuid.UUID
	Version     int64
	State       string
	Name        string
	Description string
	BannerText  string
	AutoApply   bool
	Available   bool
	Priority    int32
	StartsAt    time.Time
	EndsAt      *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Discounts   []DiscountRecord
}

type DiscountRecord struct {
	ID                      uuid.UUID
	LegacyID                uuid.UUID
	CreatedBy               uuid.UUID
	Code                    string
	Revision                int64
	VersionID               uuid.UUID
	Version                 int64
	State                   string
	Name                    string
	Description             string
	Available               bool
	Kind                    string
	BasisPoints             *int64
	FixedMinorUnits         *int64
	Currency                *string
	CurrencyRegistryVersion *int64
	Duration                string
	DurationCycles          *int64
	Priority                int32
	AllOfferings            bool
	OfferingVersionIDs      []uuid.UUID
	StableOfferingIDs       []uuid.UUID
	LegacyOfferingIDs       []uuid.UUID
	UsedCount               int64
	CampaignIDs             []uuid.UUID
	LegacyCampaignIDs       []uuid.UUID
	ValidStart              time.Time
	ValidEnd                *time.Time
	GlobalUseLimit          *int64
	PerTenantUseLimit       *int64
	CreatedAt               time.Time
	UpdatedAt               time.Time
}
