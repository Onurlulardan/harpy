package billingv1adoptionapp

import (
	"time"

	"github.com/google/uuid"
)

type SourcePlan struct {
	ID               uuid.UUID
	Name             string
	Slug             string
	Description      string
	PriceCents       int64
	AnnualPriceCents int64
	BillingInterval  string
	Currency         string
	LimitsJSON       []byte
	FeaturesJSON     []byte
	SortOrder        int32
	IsActive         bool
	IsPublic         bool
	HighlightLabel   string
	Color            string
	TrialDays        int32
	CreatedAt        *time.Time
	UpdatedAt        *time.Time
	DeletedAt        *time.Time
}

type SourceCampaign struct {
	ID          uuid.UUID
	CreatedBy   uuid.UUID
	Name        string
	Slug        string
	Description string
	StartsAt    time.Time
	EndsAt      *time.Time
	AutoApply   bool
	BannerText  string
	IsActive    bool
	CreatedAt   *time.Time
	UpdatedAt   *time.Time
	DeletedAt   *time.Time
}

type SourceCoupon struct {
	ID                uuid.UUID
	CreatedBy         uuid.UUID
	Code              string
	Name              string
	Description       string
	DiscountType      string
	DiscountValue     int64
	MaxUses           *int64
	UsedCount         int64
	MaxUsesPerTenant  int64
	ValidFrom         time.Time
	ValidUntil        *time.Time
	ApplicablePlanIDs []uuid.UUID
	Duration          string
	DurationMonths    *int64
	IsActive          bool
	CampaignID        *uuid.UUID
	CreatedAt         *time.Time
	UpdatedAt         *time.Time
	DeletedAt         *time.Time
}

type GlobalSnapshot struct {
	EnvironmentID uuid.UUID
	SnapshotAt    time.Time
	Plans         []SourcePlan
	Campaigns     []SourceCampaign
	Coupons       []SourceCoupon
}
