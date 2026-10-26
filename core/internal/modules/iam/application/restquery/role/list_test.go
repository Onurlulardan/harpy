package role

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/catalog"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type nilPort struct{}

func (*nilPort) List(context.Context, transaction.TenantScope) (catalog.Published, error) {
	return catalog.Published{}, nil
}

func TestListRequiresRealSnapshotScope(t *testing.T) {
	service, err := NewService(&nilPort{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.List(context.Background(), transaction.TenantScope{}); !errors.Is(
		err, catalog.ErrContract,
	) {
		t.Fatalf("invalid scope error = %v", err)
	}
}

func TestNewServiceRejectsTypedNilPort(t *testing.T) {
	var port *nilPort
	if _, err := NewService(port); !errors.Is(err, catalog.ErrContract) {
		t.Fatalf("typed nil error = %v", err)
	}
}
