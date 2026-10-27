package billing

import (
	catalogquery "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/catalogquery"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func globalHTTPDiscountView(snapshot catalogquery.DiscountSnapshot) billingapi.DiscountView {
	identifiers := snapshot.OfferingVersionIDs()
	offeringIDs := make([]string, len(identifiers))
	for index, identifier := range identifiers {
		offeringIDs[index] = identifier.String()
	}
	stableIdentifiers := snapshot.StableOfferingIDs()
	stableOfferingIDs := make([]string, len(stableIdentifiers))
	for index, identifier := range stableIdentifiers {
		stableOfferingIDs[index] = identifier.String()
	}
	legacyIdentifiers := snapshot.LegacyOfferingIDs()
	legacyOfferingIDs := make([]string, len(legacyIdentifiers))
	for index, identifier := range legacyIdentifiers {
		legacyOfferingIDs[index] = identifier.String()
	}
	campaignIdentifiers := snapshot.CampaignIDs()
	legacyCampaignIdentifiers := snapshot.LegacyCampaignIDs()
	var campaignID *string
	var legacyCampaignID *string
	if len(campaignIdentifiers) == 1 {
		value := campaignIdentifiers[0].String()
		campaignID = &value
	}
	if len(legacyCampaignIdentifiers) == 1 {
		value := legacyCampaignIdentifiers[0].String()
		legacyCampaignID = &value
	}
	var fixedAmount *billingapi.MoneyView
	minorUnits := snapshot.FixedMinorUnits()
	currency := snapshot.Currency()
	registryVersion := snapshot.CurrencyRegistryVersion()
	if minorUnits != nil && currency != nil && registryVersion != nil {
		fixedAmount = &billingapi.MoneyView{
			MinorUnits: *minorUnits, Currency: *currency, RegistryVersion: *registryVersion,
		}
	}
	var cycles *int
	if value := snapshot.DurationCycles(); value != nil {
		converted := int(*value)
		cycles = &converted
	}
	return billingapi.DiscountView{
		ID: snapshot.ID().String(), LegacyID: snapshot.LegacyID().String(),
		CreatedBy: snapshot.CreatedBy().String(), Code: snapshot.Code(), Name: snapshot.Name(),
		Description: snapshot.Description(), Revision: snapshot.Revision(),
		VersionID: snapshot.VersionID().String(), Version: snapshot.Version(),
		State: snapshot.State(), Kind: snapshot.Kind(),
		BasisPoints: snapshot.BasisPoints(), FixedAmount: fixedAmount,
		Duration: snapshot.Duration(), DurationCycles: cycles,
		Priority: int(snapshot.Priority()), AllOfferings: snapshot.AllOfferings(),
		OfferingIDs: offeringIDs, StableOfferingIDs: stableOfferingIDs,
		LegacyOfferingIDs: legacyOfferingIDs,
		UsedCount:         snapshot.UsedCount(), CampaignID: campaignID,
		LegacyCampaignID:     legacyCampaignID,
		AssociationAmbiguous: len(campaignIdentifiers) > 1,
		Available:            snapshot.Available(), ValidStart: formatTimestamp(snapshot.ValidStart()),
		ValidEnd:          formatOptionalTimestamp(snapshot.ValidEnd()),
		GlobalUseLimit:    snapshot.GlobalUseLimit(),
		PerTenantUseLimit: snapshot.PerTenantUseLimit(),
		CreatedAt:         formatTimestamp(snapshot.CreatedAt()),
		UpdatedAt:         formatTimestamp(snapshot.UpdatedAt()),
	}
}
