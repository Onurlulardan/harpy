package contact

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func TestUpdateContactFailsClosedWithoutDriver(t *testing.T) {
	_, _, err := (*Store)(nil).UpdateContact(context.Background(), ports.Access{}, domain.ContactUpdate{})
	if !errors.Is(err, ports.ErrUnavailable) {
		t.Fatalf("update contact error = %v, want unavailable", err)
	}
}
