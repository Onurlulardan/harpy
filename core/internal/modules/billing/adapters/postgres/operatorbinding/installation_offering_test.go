package operatorbinding

import (
	"context"
	"testing"
	"time"

	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

type offeringRow func(...any) error

func (row offeringRow) Scan(destinations ...any) error {
	return row(destinations...)
}

type offeringDatabaseFixture struct {
	row platformpostgres.Row
}

func (fixture offeringDatabaseFixture) QueryRow(
	context.Context, string, ...any,
) platformpostgres.Row {
	return fixture.row
}

func TestFetchInstallationOfferingNormalizesEffectiveTimeToUTC(t *testing.T) {
	versionID := uuid.New()
	local := time.FixedZone("local-utc", 0)
	effectiveAt := time.Date(2026, 7, 26, 1, 39, 11, 0, local)
	digest := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	database := offeringDatabase{bound: offeringDatabaseFixture{
		row: offeringRow(func(values ...any) error {
			*values[0].(*uuid.UUID), *values[1].(*int64) = versionID, 1
			*values[2].(*string), *values[3].(*time.Time) = digest, effectiveAt
			sequence, sourceDigest := int64(1), digest
			*values[4].(**int64), *values[5].(**string) = &sequence, &sourceDigest
			return nil
		}),
	}}
	result, err := database.FetchInstallationOffering(
		t.Context(), uuid.New(), "onpremise",
	)
	if err != nil || result.EffectiveAt.Location() != time.UTC ||
		!result.EffectiveAt.Equal(effectiveAt) {
		t.Fatalf("installation offering time = %s/%s, %v",
			result.EffectiveAt, result.EffectiveAt.Location(), err)
	}
}
