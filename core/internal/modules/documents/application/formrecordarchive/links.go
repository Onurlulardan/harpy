package formrecordarchive

import (
	"context"
	"slices"

	formrecordarchiveapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formrecordarchive"
	formrecordlinkapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formrecordlink"
	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/application/foundation"
	linkapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/link"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type formRecordLinks struct{ binder linkapp.StoreBinder }

func NewLinkMatcher(binder linkapp.StoreBinder) (LinkMatcher, error) {
	if foundation.NilPort(binder) {
		return nil, formrecordarchiveapi.ErrUnavailable
	}
	return &formRecordLinks{binder: binder}, nil
}

func (matcher *formRecordLinks) Match(ctx context.Context, scope platformtx.TenantScope,
	tenantID, actorID, definitionID, recordID domain.ID, itemIDs []domain.ID,
) error {
	if matcher == nil || ctx == nil || ctx.Err() != nil || !tenantID.Valid() ||
		!actorID.Valid() || !definitionID.UUIDv7() || !recordID.UUIDv7() ||
		!validItemIDs(itemIDs) || scope.OperationID() != archiveapp.OperationCreate ||
		platformtx.RequireOwner(scope, "documents") != nil {
		return formrecordarchiveapi.ErrInvalid
	}
	store, identity, err := matcher.binder.Bind(ctx, scope)
	if err != nil || store == nil || !identity.Valid() {
		return formrecordarchiveapi.ErrUnavailable
	}
	if identity.TenantID != tenantID || identity.ActorID != actorID {
		return formrecordarchiveapi.ErrForbidden
	}
	source := linkapp.Source{Owner: formrecordlinkapi.SourceOwner,
		Kind: formrecordlinkapi.SourceKind, ID: definitionID, VersionID: recordID}
	links, err := store.SourceLinks(ctx, linkapp.SourceQuery{Source: source,
		Limit: formrecordlinkapi.MaximumItemCount + 1})
	if err != nil {
		return formrecordarchiveapi.ErrUnavailable
	}
	return exactLinks(tenantID, source, links, itemIDs)
}

func exactLinks(tenantID domain.ID, source linkapp.Source,
	links []domain.Link, requested []domain.ID,
) error {
	if len(links) == 0 || len(links) != len(requested) ||
		len(links) > formrecordlinkapi.MaximumItemCount {
		return formrecordarchiveapi.ErrReferenceMismatch
	}
	items := append([]domain.ID(nil), requested...)
	slices.SortFunc(items, compareID)
	for index, link := range links {
		if link.TenantID != tenantID || link.SourceOwner != source.Owner ||
			link.SourceKind != source.Kind || link.SourceID != source.ID ||
			link.SourceVersionID != source.VersionID ||
			link.Usage != formrecordlinkapi.Usage || link.State != domain.LinkActive ||
			link.ItemID != items[index] {
			return formrecordarchiveapi.ErrReferenceMismatch
		}
	}
	return nil
}

func validItemIDs(items []domain.ID) bool {
	if len(items) == 0 || len(items) > formrecordlinkapi.MaximumItemCount {
		return false
	}
	seen := make(map[domain.ID]struct{}, len(items))
	for _, itemID := range items {
		if !itemID.UUIDv7() {
			return false
		}
		if _, duplicate := seen[itemID]; duplicate {
			return false
		}
		seen[itemID] = struct{}{}
	}
	return true
}

func compareID(left, right domain.ID) int {
	return slices.Compare(left[:], right[:])
}

var _ LinkMatcher = (*formRecordLinks)(nil)
