package discountcatalog

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	discountdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/discount"
)

func policyDigest(policy discountdomain.Policy) any {
	identifiers := policy.OfferingIDs()
	offerings := make([]string, len(identifiers))
	for index, identifier := range identifiers {
		offerings[index] = hex.EncodeToString(identifier[:])
	}
	end := int64(0)
	if value := policy.ValidEnd(); value != nil {
		end = value.UnixMicro()
	}
	basisPoints := int64(0)
	if value, ok := policy.BasisPoints(); ok {
		basisPoints = value.Value()
	}
	fixedMinor, currency, registry := int64(0), "", int64(0)
	if value, ok := policy.FixedAmount(); ok {
		fixedMinor, currency = value.MinorUnits(), value.Currency()
		registry = value.CurrencyRegistryVersion()
	}
	return struct {
		Code        string   `json:"code"`
		Kind        string   `json:"kind"`
		BasisPoints int64    `json:"basis_points"`
		FixedMinor  int64    `json:"fixed_minor"`
		Currency    string   `json:"currency"`
		Registry    int64    `json:"registry"`
		Duration    string   `json:"duration"`
		Cycles      int64    `json:"cycles"`
		Priority    int32    `json:"priority"`
		Global      *int64   `json:"global"`
		Tenant      *int64   `json:"tenant"`
		Start       int64    `json:"start"`
		End         int64    `json:"end"`
		All         bool     `json:"all"`
		Offerings   []string `json:"offerings"`
	}{string(policy.Code()), string(policy.Kind()), basisPoints, fixedMinor, currency,
		registry, string(policy.Duration().Kind()), policy.Duration().Cycles(),
		policy.Priority(), policy.GlobalLimit(), policy.TenantLimit(),
		policy.ValidStart().UnixMicro(), end, policy.AllOfferings(), offerings}
}

func presentationDigest(value discountdomain.Presentation) any {
	return struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}{value.Name(), value.Description()}
}

func campaignDigest(value CampaignUpdate) any {
	identifier := ""
	if value.CampaignID != nil {
		identifier = value.CampaignID.String()
	}
	return struct {
		Replace    bool   `json:"replace"`
		CampaignID string `json:"campaign_id"`
	}{value.Replace, identifier}
}

func requestDigest(value any) ([32]byte, error) {
	document, err := json.Marshal(value)
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(document), nil
}
