package v1adoption

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type trackingAdoptionForms struct {
	resolveScope, registerScope ports.Scope
	command                     ports.ResourceReferenceCommand
}

func (forms *trackingAdoptionForms) ResolveForm(_ context.Context, scope ports.Scope,
	_ string, _ *domain.ID, _ json.RawMessage,
) (FormReference, error) {
	forms.resolveScope = scope
	return FormReference{ID: fixtureID(8), VersionID: fixtureID(9),
		PermissionRevision: 13}, nil
}

func (forms *trackingAdoptionForms) RegisterRead(_ context.Context, scope ports.Scope,
	command ports.ResourceReferenceCommand,
) (ports.ResourceReferenceReceipt, error) {
	forms.registerScope, forms.command = scope, command
	return ports.ResourceReferenceReceipt{ReferenceID: command.Link.ID,
		OperationID: command.OperationID, TargetID: command.Link.TargetID,
		TargetVersionID: command.Link.TargetVersionID, ReferenceEpoch: 23}, nil
}

func TestCopyRegistersDerivedFormLinkForLegacyTaskInSameScope(t *testing.T) {
	now := time.Date(2026, 7, 19, 12, 0, 0, 0, time.UTC)
	boardID, columnID := fixtureID(4), fixtureID(5)
	taskID := legacyV4ID(6)
	records := []SourceRecord{
		fixtureRecord(KindBoard, boardID, LegacyBoard{ID: boardID,
			OrganizationID: fixtureID(6), CreatedBy: fixtureID(7), Name: "Forms",
			CreatedAt: now, UpdatedAt: now}),
		fixtureRecord(KindColumn, columnID, LegacyColumn{ID: columnID,
			BoardID: boardID, Name: "Open", LinkedStatus: "TODO",
			CreatedAt: now, UpdatedAt: now}),
		fixtureRecord(KindTask, taskID, LegacyTask{ID: taskID, BoardID: boardID,
			ColumnID: columnID, CreatedBy: fixtureID(7), Title: "Fill expense",
			TaskType: "FORM_FILL", Priority: "MEDIUM", Status: "TODO",
			FormSlug: "expense-request", FormFilter: json.RawMessage("{}"),
			CreatedAt: now, UpdatedAt: now}),
	}
	forms, target := &trackingAdoptionForms{}, &fixtureTarget{}
	service, err := NewService(Dependencies{Source: fixtureSource{records}, Target: target,
		Authorizer: fixtureAuthorizer{}, Forms: forms, Workflow: fixtureWorkflow{},
		Principals: fixturePrincipals{}, TZDB: fixtureTZDB{}, TZDBVersion: "2026a",
		Clock: fixtureClock{}})
	if err != nil {
		t.Fatal(err)
	}
	scope := &fixtureScope{tenant: fixtureID(2), actor: fixtureID(3)}
	if _, err = service.Copy(context.Background(), scope, fixtureRequest()); err != nil {
		t.Fatal(err)
	}
	if forms.resolveScope != scope || forms.registerScope != scope ||
		forms.command.Link.WorkResourceID != taskID ||
		forms.command.Link.ID[6]>>4 != 5 || forms.command.Link.ReferenceEpoch != 0 ||
		forms.command.Link.PermissionRevision != 13 {
		t.Fatalf("unexpected owner-scope registration: %#v", forms.command)
	}
	found := false
	for _, record := range target.plan.Records {
		if record.Kind != TargetFormLink {
			continue
		}
		var document TargetFormLinkDocument
		if json.Unmarshal(record.Document, &document) != nil ||
			document.ReferenceEpoch != 23 || record.ID != forms.command.Link.ID {
			t.Fatalf("unexpected persisted form link: %s", record.Document)
		}
		found = true
	}
	if !found {
		t.Fatal("adopted form link was not persisted")
	}
}

func legacyV4ID(seed byte) domain.ID {
	var id domain.ID
	id[0], id[6], id[8], id[15] = 1, 0x40, 0x80, seed
	return id
}
