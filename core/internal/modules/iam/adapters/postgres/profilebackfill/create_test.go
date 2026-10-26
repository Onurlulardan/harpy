package profilebackfill

import (
	"context"
	"errors"
	"testing"

	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
)

func TestCreateRejectsUnboundScope(t *testing.T) {
	_, err := (*store)(nil).Create(
		context.Background(), tenantScope{}, profilebackfillapp.PreparedPlan{},
	)
	if !errors.Is(err, profilebackfillapp.ErrInvalid) {
		t.Fatalf("Create() error = %v", err)
	}
}
