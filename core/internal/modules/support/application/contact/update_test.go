package contact

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func TestUpdateContactFailsClosedWithoutKernel(t *testing.T) {
	_, err := (*Service)(nil).UpdateContact(context.Background(), ports.Access{}, domain.ContactUpdate{})
	if !errors.Is(err, ports.ErrUnavailable) {
		t.Fatalf("update contact error = %v, want unavailable", err)
	}
}
