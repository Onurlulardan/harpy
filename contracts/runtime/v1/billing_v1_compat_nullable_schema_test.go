package runtimev1

import (
	"path/filepath"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestBillingV1CompatSchemasAcceptExplicitNullClears(t *testing.T) {
	tests := []struct {
		name     string
		schema   string
		document map[string]any
	}{
		{
			name:   "coupon",
			schema: "CouponMutation",
			document: map[string]any{
				"max_uses": nil, "valid_until": nil,
				"duration_months": nil, "campaign_id": nil,
			},
		},
		{name: "campaign", schema: "CampaignMutation", document: map[string]any{"end_date": nil}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := compileBillingV1CompatSchema(t, test.schema).Validate(test.document); err != nil {
				t.Fatalf("explicit null clear rejected: %v", err)
			}
		})
	}
}

func TestBillingV1CompatNullableSchemasRemainFailClosed(t *testing.T) {
	tests := []struct {
		name     string
		schema   string
		document map[string]any
	}{
		{name: "coupon unknown", schema: "CouponMutation", document: map[string]any{"unknown": nil}},
		{name: "coupon max uses type", schema: "CouponMutation", document: map[string]any{"max_uses": "1"}},
		{name: "coupon valid until type", schema: "CouponMutation", document: map[string]any{"valid_until": 1}},
		{name: "coupon duration type", schema: "CouponMutation", document: map[string]any{"duration_months": true}},
		{name: "coupon campaign type", schema: "CouponMutation", document: map[string]any{"campaign_id": 1}},
		{name: "campaign unknown", schema: "CampaignMutation", document: map[string]any{"unknown": nil}},
		{name: "campaign end date type", schema: "CampaignMutation", document: map[string]any{"end_date": false}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := compileBillingV1CompatSchema(t, test.schema).Validate(test.document); err == nil {
				t.Fatal("invalid document was accepted")
			}
		})
	}
}

func compileBillingV1CompatSchema(t *testing.T, name string) *jsonschema.Schema {
	t.Helper()
	const base = "https://harpy.io/contracts/http/"
	directory := filepath.Join("..", "..", "http")
	compiler := jsonschema.NewCompiler()
	for _, file := range []string{"billing-common-v2.schemas.json", "billing-v1-compat.schemas.json"} {
		value := readJSONValue(t, filepath.Join(directory, file))
		if err := compiler.AddResource(base+file, value); err != nil {
			t.Fatalf("register schema %s: %v", file, err)
		}
	}
	schema, err := compiler.Compile(base + "billing-v1-compat.schemas.json#/schemas/" + name)
	if err != nil {
		t.Fatalf("compile Billing V1 schema %s: %v", name, err)
	}
	return schema
}
