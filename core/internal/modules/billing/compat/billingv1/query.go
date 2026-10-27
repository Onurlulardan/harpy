package billingv1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func (handler *Handler) publicOfferings(
	ctx context.Context,
) ([]billingapi.OfferingView, error) {
	items, err := handler.collectCatalogPages(ctx, func(
		ctx context.Context, request billingapi.QueryRequest,
	) (billingapi.QueryPage, error) {
		return handler.public.ListPublicOfferings(ctx, billingapi.PublicQuery{
			Request: request, Transport: http.MethodGet,
		})
	})
	if err != nil {
		return nil, err
	}
	return decodeOfferings(items)
}

func (handler *Handler) platformOfferings(
	ctx context.Context,
	access billingapi.Access,
) ([]billingapi.OfferingView, error) {
	items, err := handler.collectCatalogPages(ctx, func(
		ctx context.Context, request billingapi.QueryRequest,
	) (billingapi.QueryPage, error) {
		return handler.catalog.ListOfferings(ctx, billingapi.Query{
			Access: access, Request: request, Transport: http.MethodGet,
		})
	})
	if err != nil {
		return nil, err
	}
	return decodeOfferings(items)
}

func decodeOfferings(items []json.RawMessage) ([]billingapi.OfferingView, error) {
	result := make([]billingapi.OfferingView, len(items))
	for index, item := range items {
		if err := json.Unmarshal(item, &result[index]); err != nil {
			return nil, fmt.Errorf("decode Billing offering query result: %w", err)
		}
	}
	return result, nil
}

func (handler *Handler) platformDiscounts(
	ctx context.Context,
	listAccess billingapi.Access,
	readAccess billingapi.Access,
) ([]billingapi.DiscountView, error) {
	items, err := handler.collectCatalogPages(ctx, func(
		ctx context.Context, request billingapi.QueryRequest,
	) (billingapi.QueryPage, error) {
		return handler.catalog.ListDiscounts(ctx, billingapi.Query{
			Access: listAccess, Request: request, Transport: http.MethodGet,
		})
	})
	if err != nil {
		return nil, err
	}
	identities, err := catalogRootIDs(items, "discount_id")
	if err != nil {
		return nil, err
	}
	result := make([]billingapi.DiscountView, 0, len(identities))
	for _, identity := range identities {
		view, err := handler.catalog.GetDiscount(ctx, readAccess, identity)
		if err != nil {
			return nil, err
		}
		if view.State == "active" {
			result = append(result, view)
		}
	}
	return sortDiscountViews(result)
}

func (handler *Handler) platformCampaigns(
	ctx context.Context,
	listAccess billingapi.Access,
	readAccess billingapi.Access,
) ([]billingapi.CampaignView, error) {
	items, err := handler.collectCatalogPages(ctx, func(
		ctx context.Context, request billingapi.QueryRequest,
	) (billingapi.QueryPage, error) {
		return handler.catalog.ListCampaigns(ctx, billingapi.Query{
			Access: listAccess, Request: request, Transport: http.MethodGet,
		})
	})
	if err != nil {
		return nil, err
	}
	identities, err := catalogRootIDs(items, "campaign_id")
	if err != nil {
		return nil, err
	}
	result := make([]billingapi.CampaignView, 0, len(identities))
	for _, identity := range identities {
		view, err := handler.catalog.GetCampaign(ctx, readAccess, identity)
		if err != nil {
			return nil, err
		}
		if view.State == "active" {
			result = append(result, view)
		}
	}
	return sortCampaignViews(result)
}

func findOffering(
	offerings []billingapi.OfferingView,
	id string,
) (billingapi.OfferingView, error) {
	for _, offering := range offerings {
		if offering.ID == id || offering.VersionID == id {
			return offering, nil
		}
	}
	return billingapi.OfferingView{}, billingapi.ErrNotFound
}

func monthlyPrice(offering billingapi.OfferingView) (billingapi.PriceView, error) {
	for _, price := range offering.Prices {
		if price.Cadence == "monthly" {
			return price, nil
		}
	}
	return billingapi.PriceView{}, ErrPlanChangeTarget
}

func defaultPrice(offering billingapi.OfferingView) (billingapi.PriceView, error) {
	cadence := "monthly"
	if offering.BillingInterval == "annual" {
		cadence = "yearly"
	}
	for _, price := range offering.Prices {
		if price.Cadence == cadence {
			return price, nil
		}
	}
	return billingapi.PriceView{}, ErrPlanChangeTarget
}
