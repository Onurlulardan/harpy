package licenseauthority

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

type restoreBindingStoreStub struct {
	binding RestoreBinding
}

func (store restoreBindingStoreStub) ResolveRestoreBinding(
	context.Context,
	uuid.UUID,
) (RestoreBinding, error) {
	return store.binding, nil
}

func TestResolveRestoreBindingAcceptsOneExactAuthorityHead(t *testing.T) {
	environment := uuid.New()
	expected := RestoreBinding{
		Binding: Binding{
			LicenseID: uuid.New(), LicenseeOrganizationID: uuid.New(),
			EnvironmentID: environment, DeliveryKind: DeliveryConnected,
		},
		Issuer: "forgeguard", Sequence: 4,
		SHA256:            "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		AuthorityRevision: 7,
		DatabaseNow:       time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC),
	}
	got, err := ResolveRestoreBinding(
		context.Background(), restoreBindingStoreStub{binding: expected}, environment,
	)
	if err != nil || got != expected {
		t.Fatalf("restore binding=%+v err=%v", got, err)
	}
}
