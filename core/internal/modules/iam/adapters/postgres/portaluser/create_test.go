package portaluserpostgres

import (
	"context"
	"errors"
	"testing"

	portaluser "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/portaluser"
)

func TestCreateRejectsUnboundScope(t *testing.T) {
	_, err := (&store{}).Create(context.Background(), tenantScope{}, portaluser.CreateCommand{})
	if !errors.Is(err, portaluser.ErrInvalid) {
		t.Fatalf("Create error = %v, want ErrInvalid", err)
	}
}
