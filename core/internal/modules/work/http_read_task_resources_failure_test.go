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

func TestLegacyTaskReadRejectsInconsistentRecordAndResolver(t *testing.T) {
	taskID, definitionID := workID(1), workID(2)
	versionID, recordID := workID(3), workID(4)
	raw := json.RawMessage(fmt.Sprintf(
		`[{"id":%q,"task_type":"FORM_REVIEW"}]`, taskID.String()))
	tests := []struct {
		name        string
		mutate      func(*tasks.ResourceLinkInput)
		description LegacyTaskResourceDescription
		resolverErr error
	}{
		{name: "wrong target order", mutate: func(resource *tasks.ResourceLinkInput) {
			resource.Targets[0], resource.Targets[1] = resource.Targets[1], resource.Targets[0]
		}},
		{name: "version mismatch", mutate: func(resource *tasks.ResourceLinkInput) {
			resource.Targets[1].VersionID = workID(8)
		}},
		{name: "record equals definition", mutate: func(resource *tasks.ResourceLinkInput) {
			resource.Targets[1].ID = definitionID
		}},
		{name: "resolver record mismatch", description: LegacyTaskResourceDescription{
			FormSlug: "review", FormDataID: workID(9).String()}},
		{name: "resolver unavailable", resolverErr: ports.ErrUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			document, err := parseLegacyTaskReadDocument(raw, false)
			if err != nil {
				t.Fatal(err)
			}
			resource := formResourceInput(definitionID, versionID, &recordID)
			if test.mutate != nil {
				test.mutate(resource)
			}
			description := test.description
			if description == (LegacyTaskResourceDescription{}) {
				description = LegacyTaskResourceDescription{
					FormSlug: "review", FormDataID: recordID.String()}
			}
			result, enrichErr := enrichLegacyTaskResources(context.Background(), taskScope(),
				document, map[ports.ID]*tasks.ResourceLinkInput{taskID: resource},
				&readTaskResourceResolver{description: description, err: test.resolverErr})
			if result != nil || !errors.Is(enrichErr, ports.ErrUnavailable) {
				t.Fatalf("inconsistent resource escaped: result=%s err=%v", result, enrichErr)
			}
		})
	}
}

func TestLegacyTaskReadRejectsMalformedOrDuplicateDocuments(t *testing.T) {
	taskID := workID(1).String()
	for name, raw := range map[string]json.RawMessage{
		"not array":    json.RawMessage(`{}`),
		"unknown kind": json.RawMessage(fmt.Sprintf(`[{"id":%q,"task_type":"OTHER"}]`, taskID)),
		"duplicate": json.RawMessage(fmt.Sprintf(
			`[{"id":%q,"task_type":"CUSTOM"},{"id":%q,"task_type":"CUSTOM"}]`,
			taskID, taskID)),
		"board tasks missing": json.RawMessage(`[{"board":{}}]`),
	} {
		t.Run(name, func(t *testing.T) {
			board := name == "board tasks missing"
			if _, err := parseLegacyTaskReadDocument(raw, board); !errors.Is(err, ports.ErrUnavailable) {
				t.Fatalf("malformed document accepted: %s err=%v", raw, err)
			}
		})
	}
}
