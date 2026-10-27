package query

import (
	"context"
	"errors"
	"testing"

	analyticspublication "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/publication"
	analyticsquery "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/query"
	"github.com/google/uuid"
)

func TestReadPublishedReturnsImmutableVersionAndDependencies(t *testing.T) {
	snapshot := queryFixtureSnapshot(
		analyticsquery.VersionPublished,
		analyticspublication.CacheNone,
	)
	dependency := DependencySnapshot{
		DependencyRequest: DependencyRequest{
			OwnerDomain:  "iam",
			ResourceKind: "form",
			ResourceID:   queryFixtureID,
			VersionID:    queryFixtureVersion,
		},
		CatalogVersion:       1,
		QueryContractVersion: 1,
		PolicyVersion:        1,
		OwnerGeneration:      1,
		WriterV2:             true,
		DependencySHA256:     [32]byte{1},
	}
	store := queryFixtureStore{
		latestPublished: func(context.Context, uuid.UUID) (Snapshot, error) {
			return snapshot, nil
		},
		dependencies: func(context.Context, uuid.UUID, uuid.UUID) ([]DependencySnapshot, error) {
			return []DependencySnapshot{dependency}, nil
		},
	}
	application := queryFixtureApplication()

	result, err := application.readPublishedBound(context.Background(), &store, queryFixtureID)
	if err != nil || result.Snapshot.VersionID != queryFixtureVersion ||
		len(result.Dependencies) != 1 {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestReadPublishedRejectsDraftVersion(t *testing.T) {
	store := queryFixtureStore{
		latestPublished: func(context.Context, uuid.UUID) (Snapshot, error) {
			return queryFixtureSnapshot(
				analyticsquery.VersionDraft,
				analyticspublication.CacheNone,
			), nil
		},
	}
	application := queryFixtureApplication()

	_, err := application.readPublishedBound(context.Background(), &store, queryFixtureID)
	if !errors.Is(err, ErrUnavailable) {
		t.Fatal(err)
	}
}
