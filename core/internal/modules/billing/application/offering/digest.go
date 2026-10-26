package offering

import (
	"crypto/sha256"
	"encoding/json"
	"sort"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
	offeringdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/offering"
)

type digestPrice struct {
	Cadence  string `json:"cadence"`
	Minor    int64  `json:"minor"`
	Currency string `json:"currency"`
	Registry int64  `json:"registry"`
}

type digestGrant struct {
	Key       string `json:"key"`
	Kind      string `json:"kind"`
	Boolean   *bool  `json:"boolean,omitempty"`
	Integer   *int64 `json:"integer,omitempty"`
	Unlimited bool   `json:"unlimited"`
}

func versionDigest(version offeringdomain.Version) any {
	presentation := version.Presentation()
	prices := make([]digestPrice, 0, len(version.Prices()))
	for _, price := range version.Prices() {
		amount := price.Amount()
		prices = append(prices, digestPrice{string(price.Cadence()), amount.MinorUnits(),
			amount.Currency(), amount.CurrencyRegistryVersion()})
	}
	sort.Slice(prices, func(left, right int) bool {
		if prices[left].Cadence != prices[right].Cadence {
			return prices[left].Cadence < prices[right].Cadence
		}
		return prices[left].Currency < prices[right].Currency
	})
	grants := make([]digestGrant, 0, len(version.Grants()))
	for _, grant := range version.Grants() {
		grants = append(grants, grantDigest(grant.Definition().Key(), grant.Value()))
	}
	sort.Slice(grants, func(left, right int) bool { return grants[left].Key < grants[right].Key })
	end := int64(0)
	if value := version.EffectiveEnd(); value != nil {
		end = value.UnixMicro()
	}
	return struct {
		Name        string        `json:"name"`
		Description string        `json:"description"`
		SortOrder   int32         `json:"sort_order"`
		Highlight   string        `json:"highlight_label"`
		Color       string        `json:"color"`
		Interval    string        `json:"billing_interval"`
		Public      bool          `json:"public"`
		Available   bool          `json:"available"`
		Base        bool          `json:"base"`
		Trial       int32         `json:"trial"`
		Grace       int32         `json:"grace"`
		Start       int64         `json:"start"`
		End         int64         `json:"end"`
		Prices      []digestPrice `json:"prices"`
		Grants      []digestGrant `json:"grants"`
	}{version.Name(), presentation.Description(), presentation.SortOrder(),
		presentation.HighlightLabel(), presentation.Color(),
		string(presentation.BillingInterval()), version.Public(), version.Available(),
		version.Base(), version.TrialDays(),
		version.GraceDays(), version.EffectiveAt().UnixMicro(), end, prices, grants}
}

func grantDigest(key entitlement.Key, value entitlement.Value) digestGrant {
	result := digestGrant{Key: string(key), Kind: string(value.Kind()), Unlimited: value.Unlimited()}
	if boolean, ok := value.Boolean(); ok {
		result.Boolean = &boolean
	}
	if integer, ok := value.Integer(); ok {
		result.Integer = &integer
	}
	return result
}

func requestDigest(value any) ([32]byte, error) {
	document, err := json.Marshal(value)
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(document), nil
}
