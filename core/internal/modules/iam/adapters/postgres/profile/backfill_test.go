package profile

import (
	"context"
	"errors"
	"testing"

	profilecommand "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
)

func TestBackfillRejectsInvalidCommandBeforeBinding(t *testing.T) {
	_, err := (*commandStore)(nil).Backfill(
		context.Background(), tenantScope{}, profilecommand.BackfillCommand{},
	)
	if !errors.Is(err, profilecommand.ErrInvalidCommand) {
		t.Fatalf("invalid backfill error = %v", err)
	}
}
