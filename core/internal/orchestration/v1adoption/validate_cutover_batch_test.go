package v1adoption

import (
	"encoding/json"
	"testing"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/google/uuid"
)

var cutoverOperationID = uuid.MustParse("77777777-7777-7777-8777-777777777777")

func TestValidateCutoverBatchAcceptsRecomputedApplyPlan(t *testing.T) {
	t.Parallel()
	document, payload := validApplyCutoverFixture(t)
	payload.Changes[0].After = document
	contracts := []string{
		cutoverv1.PlatformV1EnvironmentPlanContractID,
		cutoverv1.TenancyV1EnvironmentAdoptionContractID,
		cutoverv1.IAMV1ProtectedAuthorityAdoptionContractID,
		cutoverv1.BillingV1BaseAdoptionContractID,
	}
	for _, contractID := range contracts {
		contractID := contractID
		t.Run(contractID, func(t *testing.T) {
			t.Parallel()
			if err := ValidateCutoverBatch(contractID, payload); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestValidateCutoverBatchRejectsPlanAndContractTamper(t *testing.T) {
	t.Parallel()
	document, payload := validApplyCutoverFixture(t)
	var value ApplyDocument
	if err := json.Unmarshal(document, &value); err != nil {
		t.Fatal(err)
	}
	value.Plan.OwnerAuthorizationVersion++
	payload.Changes[0].After = mustJSON(t, value)
	if ValidateCutoverBatch(ApplyContractID, payload) == nil {
		t.Fatal("tampered V1 adoption plan was accepted")
	}
	if err := ValidateCutoverBatch("iam.v1_adoption.v1", payload); err != nil {
		t.Fatalf("unrelated contract was intercepted: %v", err)
	}
}

func TestValidateCutoverBatchAcceptsThirtyDayCompletion(t *testing.T) {
	t.Parallel()
	completedAt := time.Date(2026, 8, 23, 10, 0, 0, 0, time.UTC)
	value := CompletionDocument{
		ContractVersion: CompletionContractVersion, OperationID: cutoverOperationID,
		EnvironmentID: uuid.MustParse("aaaaaaaa-aaaa-7aaa-8aaa-aaaaaaaaaaaa"),
		PlanSHA256:    hash64("a"), FrontendContractSHA256: hash64("b"),
		ZeroHitSince: completedAt.Add(-30 * 24 * time.Hour),
		CompletedAt:  completedAt,
	}
	payload := atomicPayload(t, value.EnvironmentID, mustJSON(t, value))
	if err := ValidateCutoverBatch(CompletionContractID, payload); err != nil {
		t.Fatal(err)
	}
	value.LegacyBooleanHits = 1
	payload.Changes[0].After = mustJSON(t, value)
	if ValidateCutoverBatch(CompletionContractID, payload) == nil {
		t.Fatal("non-zero legacy boolean hit was accepted")
	}
}

func validApplyCutoverFixture(t *testing.T) ([]byte, cutoverv1.ChangeBatchPayload) {
	t.Helper()
	tenantID := uuid.MustParse("11111111-1111-7111-8111-111111111111")
	inventory := validInventory(ModeOnPremise, tenantID,
		uuid.MustParse("22222222-2222-7222-8222-222222222222"))
	inventory.Billing = []BillingEvidence{{
		TenantID: tenantID, Kind: BillingLicenseDerived, VerifiedV2Publication: true,
		Offering: validOffering("license_enterprise_v1", 7),
	}}
	plan, err := Evaluate(inventory)
	if err != nil {
		t.Fatal(err)
	}
	value := ApplyDocument{
		ContractVersion: ApplyContractVersion, OperationID: cutoverOperationID,
		Inventory: inventory, Plan: plan,
		AuthorityIDs: AuthorityIDs{
			SubjectID:        uuid.MustParse("88888888-8888-7888-8888-888888888881"),
			OwnerRoleID:      uuid.MustParse("88888888-8888-7888-8888-888888888882"),
			RolePermissionID: uuid.MustParse("88888888-8888-7888-8888-888888888883"),
			OwnerBindingID:   uuid.MustParse("88888888-8888-7888-8888-888888888884"),
		},
	}
	document := mustJSON(t, value)
	return document, atomicPayload(t, inventory.EnvironmentID, document)
}

func atomicPayload(t *testing.T, environmentID uuid.UUID, after []byte) cutoverv1.ChangeBatchPayload {
	t.Helper()
	return cutoverv1.ChangeBatchPayload{
		OperationID: cutoverOperationID.String(), ScopeKind: "global",
		ScopeID: environmentID.String(), CheckpointState: "complete",
		Changes: []cutoverv1.Change{{Ordinal: 1, Kind: "insert",
			Key:    mustJSON(t, map[string]string{"operation_id": cutoverOperationID.String()}),
			Before: json.RawMessage("null"), After: after}},
	}
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	document, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return document
}
