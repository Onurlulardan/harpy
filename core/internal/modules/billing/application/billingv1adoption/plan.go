package billingv1adoptionapp

import (
	"time"

	offeringdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/offering"
)

type PlanSnapshot struct {
	Name             string
	Slug             string
	Description      string
	SortOrder        int32
	HighlightLabel   string
	Color            string
	PriceCents       int64
	AnnualPriceCents int64
	BillingInterval  string
	Currency         string
	LimitsJSON       []byte
	FeaturesJSON     []byte
	IsPublic         bool
	TrialDays        int32
}

type MissingField struct {
	Document string
	Field    string
}

type PlanResult struct {
	Key           offeringdomain.Key
	Version       offeringdomain.Version
	MissingFields []MissingField
}

type ConversionRequest struct {
	Snapshot   PlanSnapshot
	IsBase     bool
	Available  bool
	SnapshotAt time.Time
}
