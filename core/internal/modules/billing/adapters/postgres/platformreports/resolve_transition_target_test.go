package platformreports

import (
	"context"
	"errors"
	"testing"

	platformreportapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/platformreports"
	"github.com/google/uuid"
)

func TestResolveTransitionTargetRejectsMissingRow(t *testing.T) {
	query := func(_ context.Context, _ string, _ ...any) row { return failedRow{} }
	_, err := (&store{query: query}).ResolveTransitionTarget(t.Context(), uuid.New())
	if !errors.Is(err, platformreportapp.ErrTarget) {
		t.Fatalf("resolve target error = %v, want target error", err)
	}
}

type failedRow struct{}

func (failedRow) Scan(...any) error { return platformreportapp.ErrTarget }
