package portaluserpostgres

import (
	"context"
	"errors"
	"testing"

	portaluser "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/portaluser"
)

func TestChangeRejectsUnboundScope(t *testing.T) {
	_, err := (&store{}).ChangeStatus(context.Background(), tenantScope{}, portaluser.RevisionCommand{})
	if !errors.Is(err, portaluser.ErrInvalid) {
		t.Fatalf("ChangeStatus error = %v, want ErrInvalid", err)
	}
}
