//go:build integration

package postgres

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
)

func verifyLeaseManifestDigestDenied(
	t *testing.T, ctx context.Context, adopter *pgx.Conn, admin *pgx.Conn,
) {
	t.Helper()
	payload := buildBoundaryPayload(t, 99, 1, 0)
	digest := fmt.Sprintf("%x", sha256.Sum256(payload))
	_, applyErr := callCompositeDocument(
		ctx, adopter, 99, "delta", payload, digest, int64(len(payload)), digestMarker('d'),
	)
	assertFinalizationError(t, applyErr, "55000", "not apply-ready")
	assertNoApplyEffects(t, ctx, admin)
}
