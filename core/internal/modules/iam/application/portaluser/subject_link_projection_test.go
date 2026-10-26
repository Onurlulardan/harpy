package portaluser

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestProjectSubjectLinksRejectsNilService(t *testing.T) {
	var service *Service
	_, err := service.ProjectSubjectLinks(
		context.Background(), platformtx.TenantScope{}, ProjectionAccess{},
		[16]byte{}, [16]byte{}, nil,
	)
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("ProjectSubjectLinks error = %v, want ErrInvalid", err)
	}
}
