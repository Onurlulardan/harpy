package adoption

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestValidateRejectsEmptyRequest(t *testing.T) {
	_, err := (&Application{}).Validate(
		context.Background(), platformtx.TenantScope{}, VerifyRequest{},
	)
	if !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
}
