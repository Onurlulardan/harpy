package folder

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

func TestCreateBoundPersistsAuditIdentityAndReplay(t *testing.T) {
	var captured Mutation
	store := fixtureStore{create: func(_ context.Context, mutation Mutation) (result domain.Folder, replay bool, err error) {
		captured = mutation
		return fixtureFolderValue(1), true, nil
	}}
	result, err := fixtureApplication().createBound(context.Background(), store, fixtureIdentity(),
		CreateRequest{DisplayName: "Contracts", IdempotencyKey: "folder-create-1"},
		fixtureFolder, fixtureNow)
	if err != nil || result.ID != fixtureFolder || captured.OperationID != fixtureOp ||
		captured.AuditID != fixtureAudit || captured.RequestSHA256 == [32]byte{} {
		t.Fatalf("result=%+v mutation=%+v err=%v", result, captured, err)
	}
}
