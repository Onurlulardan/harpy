package eventsv1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const eventSchemaBase = "https://harpy.io/contracts/events/v1/"

func TestEventSchemasCompileAndValidateContracts(t *testing.T) {
	externalCatalog := testConsumerCatalog(t, EffectExternal)
	producers := testProducerCatalog(t)
	policy := externalCatalog.Policies()[0]
	producerPolicy := producers.Policies()[0]
	envelope, _ := NewEnvelope(envelopeInput(), producers)
	tenant := eventTenantID
	bootstrap, _ := NewBootstrapReceipt(BootstrapReceipt{
		ReceiptID: receiptID, Consumer: policy.Consumer,
		ConsumerVersion: 1, TenantID: &tenant, ProducerAuthority: producerAuthority(),
		ConsumerAuthority: consumerAuthority(), AggregateType: "form_submission",
		AggregateID: aggregateID, ThroughSequence: 7, StateReference: "snapshot/submission/42",
		StateSHA256: digest("b"), CorrelationID: "bootstrap:42", CausationID: causationID,
		OccurredAt: eventTime,
	}, externalCatalog)
	effectInput := effectReceiptInput()
	effectInput.ExternalEffectReceiptID = eventString(effectID)
	effect, _ := NewEffectReceipt(effectInput, envelope, producers, externalCatalog)
	gapInput := envelopeInput()
	gapInput.EventSequence = 2
	gapEnvelope, _ := NewEnvelope(gapInput, producers)
	gap, _ := NewSequenceGapEvidence(
		gapEnvelope, policy.Consumer, policy.ConsumerVersion, 1,
		producers, externalCatalog,
	)
	operator, _ := NewOperatorReceipt(OperatorReceipt{
		ReceiptID: receiptID, Action: OperatorRetry, Consumer: policy.Consumer, ConsumerVersion: 1,
		TenantID: &tenant, ProducerAuthority: producerAuthority(), ConsumerAuthority: consumerAuthority(),
		SourceEventID: eventID, ResultEventID: eventID, SourceBusinessEffectKey: "submission/42",
		ResultBusinessEffectKey: "submission/42",
		Reason:                  "verified", TicketReference: "OPS-42", ActorID: actorID,
		CausationID: causationID, OccurredAt: eventTime,
	}, externalCatalog)
	fixtures := map[string]any{
		"integration-event.schema.json":       envelope,
		"event-producer-policy.schema.json":   producerPolicy,
		"event-consumer-policy.schema.json":   policy,
		"event-bootstrap-receipt.schema.json": bootstrap,
		"event-effect-receipt.schema.json":    effect,
		"event-operator-receipt.schema.json":  operator,
		"event-sequence-gap.schema.json":      gap,
	}
	for name, fixture := range fixtures {
		schema := compileEventSchema(t, name)
		document, err := json.Marshal(fixture)
		if err != nil {
			t.Fatalf("marshal %s: %v", name, err)
		}
		decoded, err := jsonschema.UnmarshalJSON(bytes.NewReader(document))
		if err != nil {
			t.Fatal(err)
		}
		if err := schema.Validate(decoded); err != nil {
			t.Fatalf("validate %s: %v", name, err)
		}
	}
}

func TestEventSchemaCarriesUTF8ByteAnnotations(t *testing.T) {
	value := readEventSchema(t, "event-defs.schema.json").(map[string]any)
	definitions := value["$defs"].(map[string]any)
	wanted := map[string]string{
		"domain": "63", "typeName": "128", "correlation": "64", "key": "512",
		"ticket": "256", "reason": "2048", "reference": "2048",
	}
	for name, maximum := range wanted {
		definition := definitions[name].(map[string]any)
		if fmt.Sprint(definition["x-maxUtf8Bytes"]) != maximum {
			t.Fatalf("%s UTF-8 byte annotation is missing", name)
		}
	}
	envelope := readEventSchema(t, "integration-event.schema.json").(map[string]any)
	payload := envelope["properties"].(map[string]any)["payload"].(map[string]any)
	if fmt.Sprint(payload["x-maxCanonicalUtf8Bytes"]) != "65536" ||
		payload["x-canonicalization"] != "RFC8785" {
		t.Fatal("event payload canonicalization annotations are missing")
	}
}

func compileEventSchema(t *testing.T, name string) *jsonschema.Schema {
	t.Helper()
	compiler := jsonschema.NewCompiler()
	paths, err := filepath.Glob("*.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		value := readEventSchema(t, path)
		identifier := value.(map[string]any)["$id"].(string)
		if err := compiler.AddResource(identifier, value); err != nil {
			t.Fatalf("register %s: %v", path, err)
		}
	}
	schema, err := compiler.Compile(eventSchemaBase + name)
	if err != nil {
		t.Fatalf("compile %s: %v", name, err)
	}
	return schema
}

func readEventSchema(t *testing.T, path string) any {
	t.Helper()
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	value, err := jsonschema.UnmarshalJSON(file)
	if err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return value
}
