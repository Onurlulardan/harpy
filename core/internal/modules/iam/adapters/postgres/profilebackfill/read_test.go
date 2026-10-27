package profilebackfill

import (
	"context"
	"crypto/sha256"
	"errors"
	"testing"

	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	"github.com/google/uuid"
)

func TestReadRejectsUnboundScope(t *testing.T) {
	_, err := (*store)(nil).Read(
		context.Background(), tenantScope{}, profilebackfillapp.ReadQuery{
			ResourceKind:   "user",
			SourceSchemaID: uuid.MustParse("12000000-0000-4000-8000-000000000001"),
			Limit:          profilebackfillapp.MaximumPageSize,
			Evidence: profilebackfillapp.Evidence{
				OperationID:          uuid.MustParse("12000000-0000-4000-8000-000000000002"),
				RequestSHA256:        sha256.Sum256([]byte("request")),
				ActorID:              uuid.MustParse("12000000-0000-4000-8000-000000000003"),
				ManagePermissionID:   uuid.MustParse("12000000-0000-4000-8000-000000000004"),
				ExpectedTokenVersion: 1, ExpectedAuthorizationVersion: 1,
			},
		},
	)
	if !errors.Is(err, profilebackfillapp.ErrInvalid) {
		t.Fatalf("Read() error = %v", err)
	}
}
