package schemav2

import (
	"bytes"
	"strings"
	"testing"
)

func TestPreparePlanPinsV2CanonicalIdentity(t *testing.T) {
	plan := preparationFixture(t)
	first, err := PreparePlan(plan)
	if err != nil {
		t.Fatal(err)
	}
	second, err := PreparePlan(plan)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first.CanonicalJSON(), second.CanonicalJSON()) ||
		first.SHA256() != second.SHA256() {
		t.Fatal("V2 plan preparation is not deterministic")
	}
	if first.SHA256() != "e9c765769b1ed58d8e78aa1f3de75f4194d5ccb70401384ff8a151ce98fb5a22" {
		t.Fatalf("plan digest = %s", first.SHA256())
	}
	steps := first.Steps()
	if len(steps) != 2 || steps[0].Ordinal() != 0 || steps[1].Ordinal() != 1 {
		t.Fatalf("prepared steps = %#v", steps)
	}
	if steps[1].SHA256() != "d538bb3f8b8f3ba2ae6b81669d792e2f36503cc5cd125d3d9c5d1ec7341a9446" {
		t.Fatalf("folded step digest = %s", steps[1].SHA256())
	}
	wantFragment := `"physical_name":"fld_aghup5d3dn6mbaaaaaaaaaaham_g1_folded",` +
		`"variant":"folded"`
	if !strings.Contains(string(first.CanonicalJSON()), wantFragment) {
		t.Fatalf("folded mapping is absent: %s", first.CanonicalJSON())
	}
	wrapper := steps[1].CanonicalPreimage()
	if !bytes.Contains(wrapper, []byte(`"plan_sha256":"`+first.SHA256()+`"`)) {
		t.Fatalf("step wrapper is not bound to plan: %s", wrapper)
	}
	copyBytes := wrapper
	copyBytes[0] = '['
	if steps[1].CanonicalPreimage()[0] != '{' {
		t.Fatal("prepared step aliases caller bytes")
	}
}

func TestPreparePlanRejectsWrongPairAndOrdinal(t *testing.T) {
	for name, mutate := range map[string]func(*Plan){
		"contract": func(plan *Plan) { plan.ContractVersion = "forms-schema-plan.v1" },
		"compiler": func(plan *Plan) { plan.Compiler.Version = "1.0.0" },
		"ordinal":  func(plan *Plan) { plan.Steps[1].Ordinal = 7 },
	} {
		t.Run(name, func(t *testing.T) {
			plan := preparationFixture(t)
			mutate(&plan)
			if _, err := PreparePlan(plan); err == nil {
				t.Fatal("invalid V2 plan accepted")
			}
		})
	}
}
