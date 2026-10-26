package httpapi

import "github.com/google/uuid"

func (snapshot TenantSubscriptionSnapshot) ValidForTenant(tenantID uuid.UUID) bool {
	if tenantID == uuid.Nil || snapshot.TenantID != tenantID || snapshot.ID == uuid.Nil ||
		snapshot.OfferingVersionID == uuid.Nil || snapshot.PriceVersionID == uuid.Nil ||
		!exactUUID(snapshot.View.ID, snapshot.ID) ||
		!exactUUID(snapshot.View.TenantID, tenantID) ||
		!exactUUID(snapshot.View.OfferingVersionID, snapshot.OfferingVersionID) ||
		!exactUUID(snapshot.View.PriceVersionID, snapshot.PriceVersionID) ||
		!exactUUID(snapshot.View.Offering.VersionID, snapshot.OfferingVersionID) ||
		!validOfferingIdentity(snapshot.View.Offering, snapshot.PriceVersionID) {
		return false
	}
	return true
}

func (snapshot TenantUsageSnapshot) ValidForTenant(tenantID uuid.UUID) bool {
	return tenantID != uuid.Nil && snapshot.TenantID == tenantID &&
		len(snapshot.View.Meters) <= 512
}

func (snapshot TenantDiscountSnapshot) ValidForTenant(tenantID uuid.UUID) bool {
	return tenantID != uuid.Nil && snapshot.TenantID == tenantID &&
		snapshot.DiscountID != uuid.Nil && snapshot.DiscountVersionID != uuid.Nil &&
		snapshot.SubscriptionID != uuid.Nil && snapshot.OfferingVersionID != uuid.Nil &&
		snapshot.PriceVersionID != uuid.Nil &&
		exactUUID(snapshot.View.ID, snapshot.DiscountID) &&
		exactUUID(snapshot.View.VersionID, snapshot.DiscountVersionID)
}

func validOfferingIdentity(view OfferingView, priceID uuid.UUID) bool {
	if _, valid := parseExactUUID(view.ID); !valid || len(view.Prices) < 1 ||
		len(view.Prices) > 16 || len(view.Entitlements) > 256 {
		return false
	}
	found := false
	for _, price := range view.Prices {
		identifier, valid := parseExactUUID(price.ID)
		if !valid {
			return false
		}
		found = found || identifier == priceID
	}
	return found
}

func exactUUID(raw string, expected uuid.UUID) bool {
	value, valid := parseExactUUID(raw)
	return valid && value == expected
}

func parseExactUUID(raw string) (uuid.UUID, bool) {
	value, err := uuid.Parse(raw)
	return value, err == nil && value != uuid.Nil && value.String() == raw
}
