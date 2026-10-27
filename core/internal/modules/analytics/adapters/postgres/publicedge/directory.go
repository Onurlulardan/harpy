package publicedge

import (
	"context"

	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	publicationdomain "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/publication"
)

type Directory struct{ database database }

func NewDirectory(database database) (*Directory, error) {
	if nilValue(database) {
		return nil, publicationapp.ErrInvalid
	}
	return &Directory{database: database}, nil
}

func (directory *Directory) Route(
	ctx context.Context, publicID [16]byte,
) (publicationapp.PublicRoute, error) {
	if directory == nil || directory.database == nil || ctx == nil ||
		ctx.Err() != nil || publicID == [16]byte{} {
		return publicationapp.PublicRoute{}, publicationapp.ErrInvalid
	}
	value, err := directory.database.ResolveAnalyticsPublicRoute(ctx, publicID)
	if err != nil {
		return publicationapp.PublicRoute{}, mapResolveError(err)
	}
	route := publicationapp.PublicRoute{PublicID: value.PublicID,
		TenantID: value.TenantID, ReleaseID: value.ReleaseID,
		ResourceKind:      publicationapp.ResourceKind(value.ResourceKind),
		ResourceVersionID: value.ResourceVersionID,
		Audience:          publicationdomain.Audience(value.Audience),
		State:             publicationdomain.State(value.State), Generation: uint64(value.Generation),
		CapabilityExpiresAt: value.CapabilityExpiresAt}
	if len(value.CapabilitySHA256) != 0 && len(value.CapabilitySHA256) != 32 {
		return publicationapp.PublicRoute{}, publicationapp.ErrUnavailable
	}
	copy(route.CapabilitySHA256[:], value.CapabilitySHA256)
	return route, nil
}

var _ publicationapp.PublicDirectory = (*Directory)(nil)
