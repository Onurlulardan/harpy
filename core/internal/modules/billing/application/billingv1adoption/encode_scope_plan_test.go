package billingv1adoptionapp

import (
	"testing"
	"time"
)

func TestEncodeScopePlanIsDeterministic(t *testing.T) {
	plan := emptyScopePlan("global", "70000000-0000-4000-8000-000000000007",
		"UTC", time.Date(2026, 7, 17, 1, 2, 3, 4000, time.UTC))
	evidence, err := categoryEvidence("plans", []string{}, 0)
	if err != nil {
		t.Fatal(err)
	}
	plan.Categories = append(plan.Categories, evidence)
	plan.SourceSHA256, err = sourceRoot(plan.Categories)
	if err != nil {
		t.Fatal(err)
	}
	finalizePlan(&plan)
	first, firstSHA, err := EncodeScopePlan(plan)
	if err != nil {
		t.Fatal(err)
	}
	second, secondSHA, err := EncodeScopePlan(plan)
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) || firstSHA != secondSHA || !validDigest(firstSHA) {
		t.Fatal("scope plan encoding is not deterministic")
	}
}
