package eventsv1

import (
	"bytes"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestBillingSubscriptionChangePayloadSchemaIsClosed(t *testing.T) {
	compiler := jsonschema.NewCompiler()
	resource := readEventSchema(t, "billing-subscription-change-v1.schema.json")
	identifier := resource.(map[string]any)["$id"].(string)
	if err := compiler.AddResource(identifier, resource); err != nil {
		t.Fatal(err)
	}
	schema, err := compiler.Compile(identifier)
	if err != nil {
		t.Fatal(err)
	}
	valid := []byte(`{
      "subscription_id":"018f0000-0000-7000-8000-000000000001",
      "provider_account_id":"018f0000-0000-7000-8000-000000000002",
      "offering_version_id":"018f0000-0000-7000-8000-000000000003",
      "price_version_id":"018f0000-0000-7000-8000-000000000004",
      "currency":"USD","state":"active",
      "period_start":"2026-07-17T00:00:00Z",
      "period_end":"2026-08-17T00:00:00Z",
      "trial_ends_at":null,"grace_ends_at":null,
      "cancel_at_period_end":false,"cancel_at":null,
      "revision":2,"change_kind":"provider_state_transition"
    }`)
	value, err := jsonschema.UnmarshalJSON(bytes.NewReader(valid))
	if err != nil || schema.Validate(value) != nil {
		t.Fatalf("valid billing payload was rejected: %v", err)
	}
	leaking := bytes.Replace(valid, []byte(`"revision":2`),
		[]byte(`"revision":2,"provider_secret":"leak"`), 1)
	value, err = jsonschema.UnmarshalJSON(bytes.NewReader(leaking))
	if err != nil || schema.Validate(value) == nil {
		t.Fatal("billing payload accepted an undeclared secret")
	}
}
