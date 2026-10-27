package transaction

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/google/uuid"
)

func TestBindAuthorityRejectsMissingScopeAuthorityAndWrongOwner(t *testing.T) {
	t.Parallel()
	binder, err := postgres.NewAuthorityBinder(
		"platform_async", func(postgres.BoundExecutor) *struct{} { return &struct{}{} },
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := BindAuthority(TenantScope{}, OwnerAuthority{}, binder); err !=
		postgres.ErrAuthorityBindingUnauthorized {
		t.Fatalf("missing authority error = %v", err)
	}
	wrong, err := postgres.NewAuthorityBinder(
		"forms", func(postgres.BoundExecutor) *struct{} { return &struct{}{} },
	)
	if err != nil {
		t.Fatal(err)
	}
	operation, err := NewAuthorityIssuer().Issue(
		"billing.subscription.create", "billing", 7,
		map[string]int64{"billing": 3},
	)
	if err != nil {
		t.Fatal(err)
	}
	scope := activeAuthorityScope(
		uuid.MustParse("11111111-1111-4111-8111-111111111111"), operation,
	)
	authority, err := MintOwnerAuthority(scope, "billing")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := BindAuthority(scope, authority, wrong); err !=
		postgres.ErrAuthorityBindingUnauthorized {
		t.Fatalf("wrong owner error = %v", err)
	}
}
