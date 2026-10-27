package httpapi

import "context"

type PublicQueries interface {
	ListPublicOfferings(context.Context, PublicQuery) (QueryPage, error)
	GetActiveCampaign(context.Context) (*CampaignView, error)
}
