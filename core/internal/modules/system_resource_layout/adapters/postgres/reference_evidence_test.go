package postgres

import (
	"context"
	"testing"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
	layoutapp "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/application"
	"github.com/google/uuid"
)

type referenceFixtureRow struct {
	live   int64
	digest []byte
	owners int32
}

func (row referenceFixtureRow) Scan(destinations ...any) error {
	if len(destinations) != 3 {
		return layoutUnavailable()
	}
	*destinations[0].(*int64) = row.live
	*destinations[1].(*[]byte) = append([]byte(nil), row.digest...)
	*destinations[2].(*int32) = row.owners
	return nil
}

func TestFieldReferenceEvidenceIsFailClosedAndOwnerBound(t *testing.T) {
	store := fixtureLayoutStore(referenceFixtureRow{
		live: 0, digest: append([]byte{1}, make([]byte, 31)...), owners: 1,
	})
	target := layoutapp.ReferenceTarget{
		Resource: layoutapi.ResourceUser,
		FieldID:  uuid.MustParse("33333333-3333-4333-8333-333333333333"),
	}
	evidence, err := store.FieldReferenceEvidence(context.Background(), target)
	if err != nil || evidence.Target != target || evidence.LiveReferences != 0 ||
		evidence.CheckedOwners != 1 || evidence.ManifestSHA256[0] != 1 {
		t.Fatalf("reference evidence = %+v, %v", evidence, err)
	}
	invalidStore := fixtureLayoutStore(referenceFixtureRow{
		digest: []byte{1}, owners: 1,
	})
	if _, err = invalidStore.FieldReferenceEvidence(
		context.Background(), target,
	); err != layoutapi.ErrUnavailable {
		t.Fatalf("invalid reference digest error = %v", err)
	}
}
