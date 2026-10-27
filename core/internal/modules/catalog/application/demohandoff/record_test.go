package demohandoff

import (
	"context"
	"errors"
	"testing"

	api "github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/demohandoff"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestRecordDemoHandoffRejectsUnconstructedApplication(t *testing.T) {
	var application *Application
	_, err := application.RecordDemoHandoff(
		context.Background(), platformtx.TenantScope{}, api.Request{},
	)
	if !errors.Is(err, api.ErrInvalid) {
		t.Fatalf("RecordDemoHandoff error = %v, want invalid", err)
	}
}
