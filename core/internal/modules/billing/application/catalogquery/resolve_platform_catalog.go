package catalogquery

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type PlatformCatalogSnapshot struct {
	kind     PlatformCatalogKind
	offering OfferingSnapshot
	discount DiscountSnapshot
	campaign CampaignSnapshot
}

func (snapshot PlatformCatalogSnapshot) Offering() (OfferingSnapshot, bool) {
	return snapshot.offering, snapshot.kind == PlatformOffering
}

func (snapshot PlatformCatalogSnapshot) Discount() (DiscountSnapshot, bool) {
	return snapshot.discount.clone(), snapshot.kind == PlatformDiscount
}

func (snapshot PlatformCatalogSnapshot) Campaign() (CampaignSnapshot, bool) {
	return snapshot.campaign, snapshot.kind == PlatformCampaign
}

func (queries *Queries) ResolvePlatformCatalog(
	ctx context.Context, kind PlatformCatalogKind, id uuid.UUID,
) (PlatformCatalogSnapshot, error) {
	if !queries.usable(ctx) || !validPlatformKind(kind) || !validV7(id) {
		return PlatformCatalogSnapshot{}, ErrRequest
	}
	record, found, err := queries.store.ResolvePlatformCatalog(
		ctx, PlatformResolveStoreQuery{Kind: kind, ID: id},
	)
	if err != nil {
		return PlatformCatalogSnapshot{}, fmt.Errorf("resolve platform billing catalog: %w", err)
	}
	if !found {
		return PlatformCatalogSnapshot{}, ErrNotFound
	}
	return resolvePlatformSnapshot(kind, id, record)
}

func resolvePlatformSnapshot(
	kind PlatformCatalogKind, id uuid.UUID, record PlatformCatalogRecord,
) (PlatformCatalogSnapshot, error) {
	switch {
	case kind == PlatformOffering && record.Offering != nil &&
		record.Discount == nil && record.Campaign == nil:
		snapshot, valid := projectOffering(*record.Offering, true)
		if !valid || record.Offering.ID != id || !validCatalogState(record.Offering.State) {
			return PlatformCatalogSnapshot{}, ErrResult
		}
		snapshot.state, snapshot.public = record.Offering.State, record.Offering.Public
		return PlatformCatalogSnapshot{kind: kind, offering: snapshot}, nil
	case kind == PlatformDiscount && record.Offering == nil &&
		record.Discount != nil && record.Campaign == nil:
		snapshots, valid := projectDiscounts([]DiscountRecord{*record.Discount})
		if !valid || record.Discount.ID != id || len(snapshots) != 1 ||
			record.Discount.State == "" {
			return PlatformCatalogSnapshot{}, ErrResult
		}
		return PlatformCatalogSnapshot{kind: kind, discount: snapshots[0]}, nil
	case kind == PlatformCampaign && record.Offering == nil &&
		record.Discount == nil && record.Campaign != nil:
		snapshot, valid := projectCampaign(*record.Campaign, true)
		if !valid || record.Campaign.ID != id || !validCatalogState(record.Campaign.State) {
			return PlatformCatalogSnapshot{}, ErrResult
		}
		snapshot.state = record.Campaign.State
		return PlatformCatalogSnapshot{kind: kind, campaign: snapshot}, nil
	default:
		return PlatformCatalogSnapshot{}, ErrResult
	}
}

func validPlatformKind(kind PlatformCatalogKind) bool {
	return kind == PlatformOffering || kind == PlatformDiscount || kind == PlatformCampaign
}

func validCatalogState(value string) bool { return value == "active" || value == "retired" }

func validV7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == uuid.Version(7)
}
