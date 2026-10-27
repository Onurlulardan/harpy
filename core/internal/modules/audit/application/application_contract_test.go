package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/audit/domain"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type binderProbe struct{ calls int }

func (probe *binderProbe) Bind(
	context.Context, platformtx.TenantScope, ScopeAccess,
) (Store, ScopeIdentity, error) {
	probe.calls++
	return nil, ScopeIdentity{}, ErrUnavailable
}

type authorizerProbe struct{ calls int }

func (probe *authorizerProbe) Authorize(
	context.Context, platformtx.TenantScope, Permission,
) error {
	probe.calls++
	return nil
}

type fixedIDs struct{}

func (fixedIDs) NewUUIDv7() (uuid.UUID, error) {
	return uuid.MustParse("019c0000-0000-7000-8000-000000000801"), nil
}

type fixedClock struct{}

func (fixedClock) Now() time.Time {
	return time.Date(2026, 8, 2, 14, 0, 0, 123456000, time.UTC)
}

func TestApplicationRejectsUnpinnedOperationsBeforeBinding(t *testing.T) {
	binder, authorizer := &binderProbe{}, &authorizerProbe{}
	core, err := NewApplication(binder, authorizer, fixedIDs{}, fixedClock{})
	if err != nil {
		t.Fatal(err)
	}
	_, err = core.Append(context.Background(), platformtx.TenantScope{},
		AppendIntent{OwnerOperation: "forms.record.update"})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("append error=%v", err)
	}
	_, err = core.List(context.Background(), platformtx.TenantScope{},
		domain.Filter{Page: 1, Limit: 20}, "changelog.unknown")
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("list error=%v", err)
	}
	if binder.calls != 0 || authorizer.calls != 0 {
		t.Fatalf("invalid operations reached binder=%d authorizer=%d",
			binder.calls, authorizer.calls)
	}
}

func TestApplicationIdentityAndClockSourcesRemainPinned(t *testing.T) {
	core, err := NewApplication(&binderProbe{}, &authorizerProbe{}, fixedIDs{}, fixedClock{})
	if err != nil {
		t.Fatal(err)
	}
	id, idErr := core.newID()
	now, nowErr := core.now()
	if idErr != nil || id.Version() != 7 || nowErr != nil ||
		now.Nanosecond()%1000 != 0 {
		t.Fatalf("id=%s idErr=%v now=%s nowErr=%v", id, idErr, now, nowErr)
	}
}
