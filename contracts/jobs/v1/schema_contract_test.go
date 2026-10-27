package jobsv1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const jobSchemaBase = "https://harpy.io/contracts/jobs/v1/"

func TestJobSchemasCompileAndValidateContracts(t *testing.T) {
	policy, _ := NewKindPolicy("form.export", 1, "forms", EffectExternal, 8)
	catalog := testCatalog(t, EffectExternal)
	enqueue, _ := NewEnqueue(testEnqueue(EffectExternal), catalog)
	lease, _ := NewLease(testLease(), catalog)
	attempt := baseAttempt()
	attempt.ResultState = StateSucceeded
	attempt.EffectStartedAt = stringPointer("2026-07-15T10:00:01Z")
	attempt.EffectReceiptID = stringPointer(receiptID)
	outcome, _ := NewAttemptOutcome(attempt, catalog)
	operator := operatorBase(OperatorRetry)
	operator.SourceJobID = stringPointer(jobID)
	operator.ResultJobID = stringPointer(retryJobID)
	operator.SourceEffectIdempotencyKey = stringPointer("form.export/tenant/111")
	operator.ResultEffectIdempotencyKey = stringPointer("form.export/tenant/111")
	operatorReceipt, _ := NewOperatorReceipt(operator)
	staleDisposition, _ := NewStaleJobDisposition(staleDispositionFixture(StaleJobAdopt))
	retentionRelease, _ := NewRetentionRelease(testRetentionRelease())

	fixtures := map[string]any{
		"job-kind-policy.schema.json":       policy,
		"job-enqueue.schema.json":           enqueue,
		"job-lease.schema.json":             lease,
		"job-attempt-outcome.schema.json":   outcome,
		"job-operator-receipt.schema.json":  operatorReceipt,
		"job-retention-release.schema.json": retentionRelease,
		"stale-job-disposition.schema.json": staleDisposition,
	}
	for name, fixture := range fixtures {
		schema := compileJobSchema(t, name)
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

func TestJobSchemaCarriesUTF8ByteAnnotations(t *testing.T) {
	value := readJobSchema(t, "job-defs.schema.json").(map[string]any)
	definitions := value["$defs"].(map[string]any)
	wanted := map[string]string{
		"domain": "63", "kind": "128", "code": "64", "key": "512",
		"ticket": "256", "reason": "2048", "reference": "2048",
	}
	for name, maximum := range wanted {
		definition := definitions[name].(map[string]any)
		if fmt.Sprint(definition["x-maxUtf8Bytes"]) != maximum {
			t.Fatalf("%s UTF-8 byte annotation is missing", name)
		}
	}
}

func compileJobSchema(t *testing.T, name string) *jsonschema.Schema {
	t.Helper()
	compiler := jsonschema.NewCompiler()
	paths, err := filepath.Glob("*.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range paths {
		value := readJobSchema(t, path)
		identifier := value.(map[string]any)["$id"].(string)
		if err := compiler.AddResource(identifier, value); err != nil {
			t.Fatalf("register %s: %v", path, err)
		}
	}
	schema, err := compiler.Compile(jobSchemaBase + name)
	if err != nil {
		t.Fatalf("compile %s: %v", name, err)
	}
	return schema
}

func readJobSchema(t *testing.T, path string) any {
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
