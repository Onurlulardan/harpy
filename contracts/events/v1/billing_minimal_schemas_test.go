package eventsv1

import (
	"bytes"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestBillingMinimalPayloadSchemasAreClosed(t *testing.T) {
	cases := []struct {
		name     string
		file     string
		document []byte
	}{
		{
			name: "entitlements", file: "billing-entitlements-change-v1.schema.json",
			document: []byte(`{"subscription_id":"018f0000-0000-7000-8000-000000000001","subscription_revision":1}`),
		},
		{
			name: "usage threshold", file: "billing-usage-threshold-reached-v1.schema.json",
			document: []byte(`{"meter_period_id":"018f0000-0000-7000-8000-000000000002","usage_revision":1}`),
		},
		{
			name: "invoice", file: "billing-invoice-finalized-v1.schema.json",
			document: []byte(`{"invoice_id":"018f0000-0000-7000-8000-000000000003","invoice_revision":1}`),
		},
		{
			name: "payment", file: "billing-payment-change-v1.schema.json",
			document: []byte(`{"provider_command_id":"018f0000-0000-7000-8000-000000000004","provider_command_revision":1}`),
		},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			schema := compileBillingPayloadSchema(t, test.file)
			assertBillingPayloadAccepted(t, schema, test.document)
			leaking := append(bytes.TrimSuffix(test.document, []byte("}")),
				[]byte(`,"provider_secret":"leak"}`)...)
			assertBillingPayloadRejected(t, schema, leaking)
		})
	}
}

func compileBillingPayloadSchema(t *testing.T, file string) *jsonschema.Schema {
	t.Helper()
	compiler := jsonschema.NewCompiler()
	resource := readEventSchema(t, file)
	identifier := resource.(map[string]any)["$id"].(string)
	if err := compiler.AddResource(identifier, resource); err != nil {
		t.Fatal(err)
	}
	schema, err := compiler.Compile(identifier)
	if err != nil {
		t.Fatal(err)
	}
	return schema
}

func assertBillingPayloadAccepted(
	t *testing.T,
	schema *jsonschema.Schema,
	document []byte,
) {
	t.Helper()
	value, err := jsonschema.UnmarshalJSON(bytes.NewReader(document))
	if err != nil || schema.Validate(value) != nil {
		t.Fatalf("valid billing payload was rejected: %v", err)
	}
}

func assertBillingPayloadRejected(
	t *testing.T,
	schema *jsonschema.Schema,
	document []byte,
) {
	t.Helper()
	value, err := jsonschema.UnmarshalJSON(bytes.NewReader(document))
	if err != nil || schema.Validate(value) == nil {
		t.Fatal("billing payload accepted an undeclared field")
	}
}
