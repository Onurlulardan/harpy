package publicedge

import (
	"testing"
	"time"

	publicationapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestLimiterUsesDatabaseAuthorityAndReleasesOnce(t *testing.T) {
	database := &fixtureDatabase{}
	limiter, err := NewLimiter(database, fixtureClock{now: time.Unix(1, 0).UTC()})
	if err != nil {
		t.Fatal(err)
	}
	lease, err := limiter.Admit(t.Context(), publicationapp.PublicAdmission{
		TenantID:   uuid.New(),
		ReleaseID:  uuid.MustParse("019b0000-0000-7000-8000-000000000003"),
		Generation: 1, TrustedClientIP: "127.0.0.1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = lease.Close(t.Context()); err != nil {
		t.Fatal(err)
	}
	if err = lease.Close(t.Context()); err != nil || database.releases != 1 {
		t.Fatalf("release err = %v, calls = %d", err, database.releases)
	}
	database.admitErr = &pgconn.PgError{Code: "P6105"}
	if _, err = limiter.Admit(t.Context(), publicationapp.PublicAdmission{}); err != publicationapp.ErrRateLimited {
		t.Fatalf("rate limit error = %v", err)
	}
}
