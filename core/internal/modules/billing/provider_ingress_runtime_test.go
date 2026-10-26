package billing

import (
	"context"
	"errors"
	"testing"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func TestProviderIngressRuntimeRejectsUnboundDependencies(t *testing.T) {
	store, err := NewProviderIngressStore(nil, platformasync.WriteCatalogs{})
	if store != nil || !errors.Is(err, providerapp.ErrConfiguration) {
		t.Fatalf("provider ingress store=%T error=%v", store, err)
	}
	processor, err := BindProviderWebhookProcessor(
		platformtx.JobTenantScope{}, nil, nil, nil, nil,
	)
	if processor != nil || !errors.Is(err, providerapp.ErrConfiguration) {
		t.Fatalf("provider processor=%T error=%v", processor, err)
	}
}

func TestProviderJobWriterFailsClosedWithoutBoundAuthority(t *testing.T) {
	writer := &providerJobWriter{}
	err := writer.EnqueueProviderJob(
		context.Background(), uuid.New(), providerapp.AsyncJob{},
	)
	if !errors.Is(err, providerapp.ErrWebhookDecision) {
		t.Fatalf("provider job writer error=%v", err)
	}
}
