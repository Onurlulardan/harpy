package offering

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
)

const maximumPresentationNameBytes = 128

type Version struct {
	name         string
	presentation Presentation
	public       bool
	available    bool
	base         bool
	trialDays    int32
	graceDays    int32
	effectiveAt  time.Time
	effectiveEnd *time.Time
	prices       []Price
	grants       []Grant
}

func NewVersion(
	name string,
	public bool,
	available bool,
	base bool,
	trialDays int32,
	graceDays int32,
	effectiveAt time.Time,
	effectiveEnd *time.Time,
	prices []Price,
	grants []Grant,
) (Version, error) {
	presentation, err := NewPresentation("", 0, "", "", MonthlyInterval)
	if err != nil {
		return Version{}, ErrVersion
	}
	return NewVersionWithPresentation(name, presentation, public, available, base, trialDays,
		graceDays, effectiveAt, effectiveEnd, prices, grants)
}

func NewVersionWithPresentation(
	name string,
	presentation Presentation,
	public bool,
	available bool,
	base bool,
	trialDays int32,
	graceDays int32,
	effectiveAt time.Time,
	effectiveEnd *time.Time,
	prices []Price,
	grants []Grant,
) (Version, error) {
	start, err := canonicalInstant(effectiveAt)
	if err != nil || !validName(name) || !presentation.Valid() ||
		trialDays < 0 || trialDays > 3650 ||
		graceDays < 0 || graceDays > 3650 || len(prices) == 0 || len(prices) > 16 ||
		len(grants) > 256 || !uniquePrices(prices) || !uniqueGrants(grants) {
		return Version{}, ErrVersion
	}
	end, err := canonicalEnd(effectiveEnd, start)
	if err != nil {
		return Version{}, ErrVersion
	}
	return Version{name: name, presentation: presentation, public: public,
		available: available,
		base:      base, trialDays: trialDays,
		graceDays: graceDays, effectiveAt: start, effectiveEnd: end,
		prices: append([]Price(nil), prices...), grants: append([]Grant(nil), grants...)}, nil
}

func (version Version) Name() string               { return version.name }
func (version Version) Presentation() Presentation { return version.presentation }
func (version Version) Public() bool               { return version.public }
func (version Version) Available() bool            { return version.available }
func (version Version) Base() bool                 { return version.base }
func (version Version) TrialDays() int32           { return version.trialDays }
func (version Version) GraceDays() int32           { return version.graceDays }
func (version Version) EffectiveAt() time.Time     { return version.effectiveAt }
func (version Version) Prices() []Price            { return append([]Price(nil), version.prices...) }
func (version Version) Grants() []Grant            { return append([]Grant(nil), version.grants...) }

func (version Version) EffectiveEnd() *time.Time {
	if version.effectiveEnd == nil {
		return nil
	}
	end := *version.effectiveEnd
	return &end
}

func (version Version) Valid() bool {
	if !validName(version.name) || !version.presentation.Valid() ||
		version.trialDays < 0 || version.trialDays > 3650 ||
		version.graceDays < 0 || version.graceDays > 3650 ||
		len(version.prices) == 0 || len(version.prices) > 16 || len(version.grants) > 256 ||
		!uniquePrices(version.prices) || !uniqueGrants(version.grants) {
		return false
	}
	start, err := canonicalInstant(version.effectiveAt)
	if err != nil {
		return false
	}
	_, err = canonicalEnd(version.effectiveEnd, start)
	return err == nil
}

func validName(value string) bool {
	return utf8.ValidString(value) && strings.TrimSpace(value) == value && value != "" &&
		len(value) <= maximumPresentationNameBytes
}

func uniquePrices(prices []Price) bool {
	seen := map[string]struct{}{}
	for _, price := range prices {
		if !validCadence(price.cadence) || price.amount.MinorUnits() < 0 {
			return false
		}
		key := string(price.cadence) + "\x00" + price.amount.Currency()
		if _, exists := seen[key]; exists {
			return false
		}
		seen[key] = struct{}{}
	}
	return true
}

func uniqueGrants(grants []Grant) bool {
	seen := map[entitlement.Key]struct{}{}
	for _, grant := range grants {
		if !grant.definition.Accepts(grant.value) {
			return false
		}
		key := grant.definition.Key()
		if _, exists := seen[key]; exists {
			return false
		}
		seen[key] = struct{}{}
	}
	return true
}
