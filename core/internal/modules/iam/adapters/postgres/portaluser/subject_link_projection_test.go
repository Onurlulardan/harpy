package portaluserpostgres

import (
	"context"
	"errors"
	"testing"

	portaluser "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/portaluser"
)

func TestProjectSubjectLinksRejectsUnboundScope(t *testing.T) {
	_, err := (&store{}).ProjectSubjectLinks(
		context.Background(), tenantScope{}, portaluser.ProjectionAccess{},
		[16]byte{}, [16]byte{}, nil,
	)
	if !errors.Is(err, portaluser.ErrInvalid) {
		t.Fatalf("ProjectSubjectLinks error = %v, want ErrInvalid", err)
	}
}
