//go:build integration

package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
)

func verifyNullChangeKindDenied(
	t *testing.T, ctx context.Context, adopter *pgx.Conn, admin *pgx.Conn,
) {
	t.Helper()
	payload := withNullChangeKind(t, buildBoundaryPayload(t, 98, 1, 0))
	digest := fmt.Sprintf("%x", sha256.Sum256(payload))
	_, err := callCompositeDocument(
		ctx, adopter, 98, "delta", payload, digest, int64(len(payload)), digestMarker('a'),
	)
	assertFinalizationError(t, err, "22023", "batch shape differs")
	assertNoApplyEffects(t, ctx, admin)
}

func withNullChangeKind(t *testing.T, document []byte) []byte {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(document, &payload); err != nil {
		t.Fatal(err)
	}
	payload["changes"].([]any)[0].(map[string]any)["kind"] = nil
	mutated, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	return mutated
}
