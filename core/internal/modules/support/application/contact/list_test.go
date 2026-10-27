package contact

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func TestListContactsFailsClosedWithoutKernel(t *testing.T) {
	_, err := (*Service)(nil).ListContacts(context.Background(), ports.Access{}, domain.ContactFilter{})
	if !errors.Is(err, ports.ErrUnavailable) {
		t.Fatalf("list contacts error = %v, want unavailable", err)
	}
}
