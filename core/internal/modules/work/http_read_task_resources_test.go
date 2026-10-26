package work

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application/tasks"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type readTaskResourceResolver struct {
	description LegacyTaskResourceDescription
	err         error
	scope       ports.Scope
	calls       int
}

func (*readTaskResourceResolver) Resolve(
	context.Context, ports.Scope, LegacyTaskResourceRequest,
) (*tasks.ResourceLinkInput, error) {
	return nil, ports.ErrUnavailable
}

func (fixture *readTaskResourceResolver) Describe(
	_ context.Context, scope ports.Scope, _ *tasks.ResourceLinkInput,
) (LegacyTaskResourceDescription, error) {
	fixture.calls++
	fixture.scope = scope
	return fixture.description, fixture.err
}

func TestLegacyTaskReadEnrichesListFromCanonicalTargets(t *testing.T) {
	taskID, definitionID := workID(1), workID(2)
	versionID, recordID := workID(3), workID(4)
	raw := json.RawMessage(fmt.Sprintf(
		`[{"id":%q,"task_type":"FORM_FILL","position":9007199254740993},`+
			`{"id":%q,"task_type":"CUSTOM"}]`, taskID.String(), workID(5).String()))
	document, err := parseLegacyTaskReadDocument(raw, false)
	if err != nil {
		t.Fatal(err)
	}
	scope := taskScope()
	resolver := &readTaskResourceResolver{description: LegacyTaskResourceDescription{
		FormSlug: "leave-request", FormDataID: recordID.String()}}
	result, err := enrichLegacyTaskResources(context.Background(), scope, document,
		map[ports.ID]*tasks.ResourceLinkInput{taskID: formResourceInput(
			definitionID, versionID, &recordID)}, resolver)
	if err != nil || resolver.calls != 1 || resolver.scope != scope {
		t.Fatalf("enrichment failed: calls=%d err=%v", resolver.calls, err)
	}
	var values []map[string]json.RawMessage
	if json.Unmarshal(result, &values) != nil || len(values) != 2 ||
		string(values[0]["form_slug"]) != `"leave-request"` ||
		string(values[0]["form_data_id"]) != `"`+recordID.String()+`"` ||
		string(values[0]["position"]) != "9007199254740993" ||
		string(values[1]["form_slug"]) != `""` ||
		string(values[1]["form_data_id"]) != "null" {
		t.Fatalf("legacy task aliases drifted: %s", result)
	}
}

func TestLegacyBoardDetailEnrichesNestedDefinitionOnlyTask(t *testing.T) {
	taskID, definitionID, versionID := workID(1), workID(2), workID(3)
	raw := json.RawMessage(fmt.Sprintf(
		`[{"board":{},"columns":[],"tasks":{"column":[`+
			`{"id":%q,"task_type":"FORM_REVIEW"}]}}]`, taskID.String()))
	document, err := parseLegacyTaskReadDocument(raw, true)
	if err != nil {
		t.Fatal(err)
	}
	resolver := &readTaskResourceResolver{description: LegacyTaskResourceDescription{
		FormSlug: "expense-review"}}
	result, err := enrichLegacyTaskResources(context.Background(), taskScope(), document,
		map[ports.ID]*tasks.ResourceLinkInput{taskID: formResourceInput(
			definitionID, versionID, nil)}, resolver)
	if err != nil || resolver.calls != 1 {
		t.Fatalf("board enrichment failed: calls=%d err=%v", resolver.calls, err)
	}
	var values []struct {
		Tasks map[string][]map[string]json.RawMessage `json:"tasks"`
	}
	if json.Unmarshal(result, &values) != nil || len(values) != 1 ||
		string(values[0].Tasks["column"][0]["form_slug"]) != `"expense-review"` ||
		string(values[0].Tasks["column"][0]["form_data_id"]) != "null" {
		t.Fatalf("board aliases drifted: %s", result)
	}
}

func TestLegacyTaskReadFailsClosedWithoutExactResource(t *testing.T) {
	taskID := workID(1)
	raw := json.RawMessage(fmt.Sprintf(
		`[{"id":%q,"task_type":"FORM_FILL"}]`, taskID.String()))
	for name, resources := range map[string]map[ports.ID]*tasks.ResourceLinkInput{
		"missing": {},
		"custom owner": {taskID: {Owner: ports.SourceWorkflow,
			Targets: []tasks.ResourceTargetInput{{Kind: "FORM", ID: workID(3),
				VersionID: workID(4)}}}},
	} {
		t.Run(name, func(t *testing.T) {
			document, parseErr := parseLegacyTaskReadDocument(raw, false)
			if parseErr != nil {
				t.Fatal(parseErr)
			}
			result, err := enrichLegacyTaskResources(context.Background(), taskScope(),
				document, resources, &readTaskResourceResolver{})
			if result != nil || !errors.Is(err, ports.ErrUnavailable) {
				t.Fatalf("corrupt resource escaped: result=%s err=%v", result, err)
			}
		})
	}
}
