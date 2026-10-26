package portaluser

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestReadRejectsNilService(t *testing.T) {
	var service *Service
	_, err := service.Read(context.Background(), platformtx.TenantScope{}, [16]byte{})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("Read error = %v, want ErrInvalid", err)
	}
}
