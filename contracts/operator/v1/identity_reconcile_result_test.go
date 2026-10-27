package operatorv1

import (
	"bytes"
	"strings"
	"testing"
)

func TestIdentityReconcileResultIsExactAndNonSecret(t *testing.T) {
	result := IdentityReconcileResult{ContractVersion: IdentityReconcileResultContract,
		Status: "completed", RestoreOperationID: "11111111-1111-4111-8111-111111111111",
		EnvironmentID:      "22222222-2222-4222-8222-222222222222",
		TenantID:           "33333333-3333-4333-8333-333333333333",
		AdmissionState:     "identity_reactivation_required",
		InvalidationSHA256: strings.Repeat("a", 64), CustodyGeneration: 2,
		CustodySHA256: strings.Repeat("b", 64)}
	var output bytes.Buffer
	if err := WriteIdentityReconcileResult(&output, result); err != nil {
		t.Fatal(err)
	}
	text := output.String()
	if strings.Contains(text, "security_epoch") || strings.Contains(text, "password") ||
		!strings.Contains(text, IdentityReconcileResultContract) {
		t.Fatalf("unsafe result: %s", text)
	}
	result.Status = "failed"
	if WriteIdentityReconcileResult(&output, result) == nil {
		t.Fatal("invalid result status was accepted")
	}
}

func TestIdentityReconcileResultReaderRejectsNonExactDocuments(t *testing.T) {
	valid := `{"contract_version":"operator-identity-reconcile-result.v1",` +
		`"status":"completed","restore_operation_id":"11111111-1111-4111-8111-111111111111",` +
		`"environment_id":"22222222-2222-4222-8222-222222222222",` +
		`"tenant_id":"33333333-3333-4333-8333-333333333333",` +
		`"admission_state":"identity_reactivation_required",` +
		`"invalidation_sha256":"` + strings.Repeat("a", 64) + `",` +
		`"custody_generation":2,"custody_result_sha256":"` + strings.Repeat("b", 64) + `"}`
	result, err := ReadIdentityReconcileResult(strings.NewReader(valid))
	if err != nil || result.EnvironmentID != "22222222-2222-4222-8222-222222222222" {
		t.Fatalf("valid result = %#v, %v", result, err)
	}
	for _, document := range []string{valid + `{}`, strings.Replace(valid, `}`, `,"extra":true}`, 1)} {
		if _, err = ReadIdentityReconcileResult(strings.NewReader(document)); err == nil {
			t.Fatalf("non-exact result was accepted: %s", document)
		}
	}
}
