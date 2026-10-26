package demomaterialization

import (
	"context"
	"errors"
	"testing"

	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestTargetRejectsUnboundScope(t *testing.T) {
	application := &Application{}
	component := fixtureRequest(t).Component
	component.Generation = 0
	_, err := application.Target(
		context.Background(), platformtx.TenantScope{}, component,
	)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("unbound target error=%v", err)
	}
}

func TestTargetRejectsForeignOwner(t *testing.T) {
	application := &Application{}
	component := contract.Component{
		ID: testComponent, Owner: "forms", Contract: catalogContract(),
		ContentSHA256: fixtureRequest(t).Component.ContentSHA256,
	}
	_, err := application.Target(
		context.Background(), platformtx.TenantScope{}, component,
	)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("foreign owner target error=%v", err)
	}
}
