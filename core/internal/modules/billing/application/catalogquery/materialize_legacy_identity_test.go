package catalogquery

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestMaterializeLegacyIdentitySupportsAdoptedAndCanonicalInputs(t *testing.T) {
	legacy := uuid.MustParse("74000000-0000-4000-8000-000000000074")
	canonical := uuid.MustParse("019c92a0-0000-7000-8000-000000000074")
	queries, _ := New(fakeStore{legacy: func(
		_ context.Context, query LegacyIdentityStoreQuery,
	) (LegacyIdentityRecord, bool, error) {
		if query.Kind != PlatformDiscount ||
			query.LegacyID != legacy && query.LegacyID != canonical {
			t.Fatalf("query=%#v", query)
		}
		return LegacyIdentityRecord{
			LegacyID: legacy, CanonicalID: canonical, Kind: PlatformDiscount,
		}, true, nil
	}})
	for _, input := range []uuid.UUID{legacy, canonical} {
		result, err := queries.MaterializeLegacyIdentity(
			context.Background(), PlatformDiscount, input,
		)
		if err != nil || result.LegacyID != legacy || result.CanonicalID != canonical {
			t.Fatalf("input=%s result=%#v err=%v", input, result, err)
		}
	}
}

func TestMaterializeLegacyIdentityFailsClosed(t *testing.T) {
	legacy := uuid.MustParse("74000000-0000-4000-8000-000000000074")
	tests := []struct {
		name  string
		store fakeStore
		want  error
	}{
		{name: "missing", store: fakeStore{legacy: func(
			context.Context, LegacyIdentityStoreQuery,
		) (LegacyIdentityRecord, bool, error) {
			return LegacyIdentityRecord{}, false, nil
		}}, want: ErrNotFound},
		{name: "wrong kind", store: fakeStore{legacy: func(
			context.Context, LegacyIdentityStoreQuery,
		) (LegacyIdentityRecord, bool, error) {
			return LegacyIdentityRecord{
				LegacyID:    legacy,
				CanonicalID: uuid.MustParse("019c92a0-0000-7000-8000-000000000074"),
				Kind:        PlatformDiscount,
			}, true, nil
		}}, want: ErrRequest},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			queries, _ := New(test.store)
			_, err := queries.MaterializeLegacyIdentity(
				context.Background(), PlatformCampaign, legacy,
			)
			if !errors.Is(err, test.want) {
				t.Fatalf("error=%v want=%v", err, test.want)
			}
		})
	}
}
