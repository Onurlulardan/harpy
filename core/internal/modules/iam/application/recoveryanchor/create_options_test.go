package recoveryanchor

import (
	"context"
	"errors"
	"testing"

	anchordomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestCreateOptionsRejectsMissingService(t *testing.T) {
	var service *Service
	var scope platformtx.TenantScope
	_, err := service.CreateOptions(
		context.Background(),
		scope,
		anchordomain.OptionsCommand{},
	)
	if !errors.Is(err, anchordomain.ErrInvalid) {
		t.Fatalf("CreateOptions() error = %v, want ErrInvalid", err)
	}
}
