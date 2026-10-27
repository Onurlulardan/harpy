package portaluserpostgres

import (
	"context"
	"errors"
	"testing"

	portaluser "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/portaluser"
)

func TestReadRejectsUnboundScope(t *testing.T) {
	_, err := (&store{}).Read(context.Background(), tenantScope{}, [16]byte{})
	if !errors.Is(err, portaluser.ErrInvalid) {
		t.Fatalf("Read error = %v, want ErrInvalid", err)
	}
}
