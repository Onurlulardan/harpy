package recovery

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestRecordCopyDerivesCanonicalRecoveryObjectKey(t *testing.T) {
	fixture := &storeFixture{}
	service := newFixtureService(fixture)
	setID := uuid.MustParse("30000000-0000-4000-8000-000000000003")
	versionID := uuid.MustParse("40000000-0000-4000-8000-000000000004")
	item := InventoryItem{
		VersionID: versionID, ObjectKey: "source", ObjectVersion: "v1",
		ObjectGeneration: 1, CiphertextSHA256: [32]byte{1},
		CiphertextBytes: 1, KeyID: "key",
	}
	if err := service.RecordCopy(context.Background(), setID, item, "v2"); err != nil {
		t.Fatalf("record copy: %v", err)
	}
	wantSuffix := "/documents/tenants/20000000-0000-4000-8000-000000000002" +
		"/versions/" + versionID.String() + ".fcdoc1"
	if !strings.HasSuffix(fixture.copyReceipt.TargetObjectKey, wantSuffix) {
		t.Fatalf("unexpected canonical key: %s", fixture.copyReceipt.TargetObjectKey)
	}
}
