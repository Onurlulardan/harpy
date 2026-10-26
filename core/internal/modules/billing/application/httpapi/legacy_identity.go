package httpapi

import "context"

type LegacyIdentityKind string

const (
	LegacyIdentityOffering     LegacyIdentityKind = "offering"
	LegacyIdentityDiscount     LegacyIdentityKind = "discount"
	LegacyIdentityCampaign     LegacyIdentityKind = "campaign"
	LegacyIdentitySubscription LegacyIdentityKind = "subscription"
)

type LegacyIdentityQuery struct {
	Access   Access
	Kind     LegacyIdentityKind
	LegacyID string
}

type LegacyIdentityView struct {
	LegacyID    string `json:"legacy_id"`
	CanonicalID string `json:"canonical_id"`
}

type LegacyIdentityResolver interface {
	ResolveLegacyIdentity(context.Context, LegacyIdentityQuery) (LegacyIdentityView, error)
}

type PublicLegacyIdentityResolver interface {
	LegacyIdentityResolver
	ResolvePublicLegacyIdentity(
		context.Context, LegacyIdentityKind, string,
	) (LegacyIdentityView, error)
}
