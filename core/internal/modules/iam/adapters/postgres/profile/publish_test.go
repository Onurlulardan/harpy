package profile

import (
	"context"
	"errors"
	"testing"

	profilecommand "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
)

func TestPublishRejectsInvalidCommandBeforeBinding(t *testing.T) {
	_, err := (*commandStore)(nil).Publish(
		context.Background(), tenantScope{}, profilecommand.PublishCommand{},
	)
	if !errors.Is(err, profilecommand.ErrInvalidCommand) {
		t.Fatalf("invalid publish error = %v", err)
	}
}
