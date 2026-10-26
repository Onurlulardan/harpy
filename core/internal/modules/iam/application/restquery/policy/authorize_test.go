package policy

import (
	"context"
	"errors"
	"testing"

	restauthorization "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/authorization"
	iamauthorization "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authorization"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type testPorts struct{}

func (*testPorts) Authorize(
	context.Context,
	transaction.TenantScope,
	iamauthorization.Request,
) (iamauthorization.Decision, error) {
	return iamauthorization.Decision{}, catalog.ErrContract
}

func (*testPorts) Catalog(
	context.Context,
	transaction.TenantScope,
	profile.ResourceKind,
) (catalog.Published, error) {
	return catalog.Published{}, catalog.ErrContract
}

func (*testPorts) Versions(
	context.Context,
	transaction.TenantScope,
	uuid.UUID,
) (ActorVersions, error) {
	return ActorVersions{}, catalog.ErrContract
}

func TestAuthorizeRequiresRealSnapshotScope(t *testing.T) {
	ports := &testPorts{}
	service, err := NewService(ports, ports, ports)
	if err != nil {
		t.Fatal(err)
	}
	_, err = service.EvaluateQuery(
		context.Background(), transaction.TenantScope{}, restauthorization.PolicyRequest{},
	)
	if !errors.Is(err, catalog.ErrContract) {
		t.Fatalf("invalid policy request error = %v", err)
	}
}

func TestClassificationActionsAreClosed(t *testing.T) {
	tests := map[profile.Classification]string{
		profile.ClassificationPublic: "read", profile.ClassificationInternal: "read",
		profile.ClassificationConfidential: "read.confidential",
		profile.ClassificationRestricted:   "read.restricted",
	}
	for classification, expected := range tests {
		actual, err := classificationAction(classification)
		if err != nil || actual != expected {
			t.Fatalf("action %q = %q, %v", classification, actual, err)
		}
	}
	if _, err := classificationAction(profile.Classification("future")); !errors.Is(
		err, catalog.ErrContract,
	) {
		t.Fatal("future classification was accepted")
	}
}
