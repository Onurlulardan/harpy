package recoveryset

import (
	"fmt"
	"testing"
	"time"
)

func TestApplyRetentionKeepsTwoActiveFiveWeeklyAndThreeMonthly(t *testing.T) {
	start := time.Date(2026, 7, 23, 0, 0, 0, 0, time.UTC)
	values := make([]RetentionCandidate, 0, 16)
	for index := 0; index < 16; index++ {
		values = append(values, RetentionCandidate{
			ID: string(rune('a' + index)), ProtectedAt: start.AddDate(0, 0, -8*index)})
	}
	decision := ApplyRetention(values)
	if len(decision.Keep) != 10 || len(decision.Delete) != 6 {
		t.Fatalf("retention keep/delete = %d/%d, want 10/6: %#v",
			len(decision.Keep), len(decision.Delete), decision)
	}
	if decision.Keep[0] != "a" || decision.Keep[1] != "b" {
		t.Fatalf("two newest active sets were not retained: %#v", decision.Keep)
	}
}

func TestRetentionKeepsLegalHoldOutsidePolicySlots(t *testing.T) {
	base := time.Date(2026, 7, 23, 0, 0, 0, 0, time.UTC)
	values := []RetentionCandidate{{ID: "hold", ProtectedAt: base.Add(-365 * 24 * time.Hour), LegalHold: true}}
	for index := 0; index < 15; index++ {
		values = append(values, RetentionCandidate{ID: fmt.Sprintf("set-%02d", index),
			ProtectedAt: base.Add(-time.Duration(index) * 8 * 24 * time.Hour)})
	}
	decision := ApplyRetention(values)
	if !contains(decision.Keep, "hold") {
		t.Fatal("legal hold was selected for deletion")
	}
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}

func TestModeSelectionIsExactXORAndDisablesSaaSDatabaseRecovery(t *testing.T) {
	onPremise := ModeSelection{Mode: ModeOnPremise, OnPremiseStorage: true}
	saas := ModeSelection{Mode: ModeSaaS, SaaSProviderRegistry: true}
	if onPremise.Validate() != nil || !onPremise.AllowsDatabaseRecovery() ||
		saas.Validate() != nil || saas.AllowsDatabaseRecovery() {
		t.Fatal("mode selection contract drifted")
	}
	invalid := []ModeSelection{{Mode: ModeOnPremise},
		{Mode: ModeOnPremise, OnPremiseStorage: true, SaaSProviderRegistry: true},
		{Mode: ModeSaaS, OnPremiseStorage: true}, {Mode: ModeSaaS}}
	for _, value := range invalid {
		if value.Validate() == nil {
			t.Fatalf("invalid mode selection was accepted: %#v", value)
		}
	}
}
