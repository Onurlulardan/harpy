package profile

import (
	"context"
	"errors"
	"testing"

	profilecommand "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
)

func TestUpdateSelfRejectsInvalidCommandBeforeBinding(t *testing.T) {
	_, err := (*commandStore)(nil).UpdateSelf(
		context.Background(), tenantScope{}, profilecommand.SelfUpdateCommand{},
	)
	if !errors.Is(err, profilecommand.ErrInvalidCommand) {
		t.Fatalf("invalid self update error = %v", err)
	}
}
