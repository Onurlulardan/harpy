package onboarding

import (
	"context"
	"errors"
	"testing"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestClaimClassifiesLeaseAndRevisionFences(t *testing.T) {
	for _, candidate := range []struct {
		code string
		want error
	}{{"55P03", tenantapi.ErrOnboardingLeaseHeld},
		{"40001", tenantapi.ErrOnboardingConflict}} {
		storage := &postgresStore{execute: noopExecute, queryRow: func(
			context.Context, string, ...any,
		) Row {
			return rowFixture(func(...any) error {
				return &pgconn.PgError{Code: candidate.code}
			})
		}}
		_, err := storage.Claim(context.Background(), tenantapi.ClaimAttemptRequest{
			OperationID: uuid.New(), AttemptID: uuid.New(), LeaseOwner: "operator",
			LeaseTokenSHA256: string(make([]byte, 64)), ExpectedRevision: 1,
			LeaseDuration: time.Minute,
		})
		if !errors.Is(err, candidate.want) {
			t.Fatalf("code %s error = %v", candidate.code, err)
		}
	}
}
