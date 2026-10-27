package formrecordarchive

import (
	"errors"
	"testing"

	formrecordarchiveapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formrecordarchive"
	formrecordlinkapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formrecordlink"
	linkapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/link"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func TestExactLinksRejectsItemAndTenantSubstitution(t *testing.T) {
	tenantID := fixtureDomainID("018f0000-0000-7000-8000-000000000181")
	definitionID := fixtureDomainID("018f0000-0000-7000-8000-000000000182")
	recordID := fixtureDomainID("018f0000-0000-7000-8000-000000000183")
	first := fixtureDomainID("018f0000-0000-7000-8000-000000000184")
	second := fixtureDomainID("018f0000-0000-7000-8000-000000000185")
	source := linkapp.Source{Owner: formrecordlinkapi.SourceOwner,
		Kind: formrecordlinkapi.SourceKind, ID: definitionID, VersionID: recordID}
	links := []domain.Link{
		linkedItem(tenantID, source, first, 1), linkedItem(tenantID, source, second, 2),
	}
	if err := exactLinks(tenantID, source, links, []domain.ID{second, first}); err != nil {
		t.Fatalf("exact reordered request was rejected: %v", err)
	}
	foreign := append([]domain.Link(nil), links...)
	foreign[0].TenantID = fixtureDomainID("018f0000-0000-7000-8000-000000000186")
	if err := exactLinks(tenantID, source, foreign, []domain.ID{first, second}); !errors.Is(err, formrecordarchiveapi.ErrReferenceMismatch) {
		t.Fatalf("foreign tenant link error = %v", err)
	}
	if err := exactLinks(tenantID, source, links, []domain.ID{first,
		fixtureDomainID("018f0000-0000-7000-8000-000000000187")}); !errors.Is(err, formrecordarchiveapi.ErrReferenceMismatch) {
		t.Fatalf("substituted item error = %v", err)
	}
}

func linkedItem(tenantID domain.ID, source linkapp.Source,
	itemID domain.ID, seed byte,
) domain.Link {
	return domain.Link{ID: domain.ID{seed}, TenantID: tenantID, ItemID: itemID,
		SourceOwner: source.Owner, SourceKind: source.Kind, SourceID: source.ID,
		SourceVersionID: source.VersionID, Usage: formrecordlinkapi.Usage,
		State: domain.LinkActive}
}
