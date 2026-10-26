package customerprovisioning

import (
	"context"
	"errors"
	"testing"

	provisioning "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/application/customerprovisioning"
)

func TestReadRejectsMissingIdentifiers(t *testing.T) {
	store := &Store{}
	if _, err := store.Read(context.Background(), [16]byte{}, [16]byte{}); !errors.Is(err, provisioning.ErrInvalid) {
		t.Fatalf("Read() error = %v, want %v", err, provisioning.ErrInvalid)
	}
}
