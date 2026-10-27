package systemresource

import "testing"

func TestActorEvidenceRequiresCurrentVersionEvidence(t *testing.T) {
	tenantID := mustSystemResourceID(t, "018f0f4e-7b61-7c25-8000-000000000011")
	actorID := mustSystemResourceID(t, "018f0f4e-7b61-7c25-8000-000000000012")
	evidence, err := NewActorEvidence(tenantID, actorID, 3, 5)
	actualTenant, tenantErr := evidence.TenantID()
	actualActor, actorErr := evidence.ActorID()
	tokenVersion, authorizationVersion, versionsErr := evidence.Versions()
	if err != nil || tenantErr != nil || actorErr != nil || versionsErr != nil ||
		actualTenant != tenantID || actualActor != actorID || tokenVersion != 3 ||
		authorizationVersion != 5 {
		t.Fatalf("evidence mismatch: %+v %v", evidence, err)
	}
	if _, err := NewActorEvidence(tenantID, actorID, 0, 5); err == nil {
		t.Fatal("zero token version accepted")
	}
}
