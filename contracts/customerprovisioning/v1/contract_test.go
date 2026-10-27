package customerprovisioningv1

import (
	"testing"

	"github.com/google/uuid"
)

func TestProvisionPurposeBindsRequestAndIdempotency(t *testing.T) {
	t.Parallel()
	tenant := uuid.MustParse("11111111-1111-7111-8111-111111111111")
	first, err := ProvisionStepUpPurpose(tenant, "customer-acme",
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil || first == [32]byte{} {
		t.Fatalf("purpose = %x, err=%v", first, err)
	}
	second, _ := ProvisionStepUpPurpose(tenant, "customer-other",
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if first == second {
		t.Fatal("idempotency substitution did not change purpose")
	}
	third, _ := ProvisionStepUpPurpose(tenant, "customer-acme",
		"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	if first == third {
		t.Fatal("request substitution did not change purpose")
	}
}
